#!/usr/bin/env python3
"""Explicit, offline Homebrew acceptance using only local test fixtures.

Not part of the automated Go suite. Requires an existing Homebrew source checkout
with tag 7.0.10 and its already-installed matching portable Ruby. No downloads,
external taps, dependency installs, sudo, host package updates, or publication.
"""
import argparse
import hashlib
import io
import json
import os
import platform
from pathlib import Path
import shutil
import subprocess
import sys
import tarfile
from native_brew_paths import validate_root

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--root', required=True, type=Path, help='new dedicated test directory')
parser.add_argument('--brew-source', required=True, type=Path, help='existing Homebrew Git checkout')
parser.add_argument('--upmacx', required=True, type=Path, help='locally built Apple Silicon arm64 CLI')
parser.add_argument("--casks-only", action="store_true", help="validate casks independently when formula builds are blocked")
a = parser.parse_args()
if sys.platform != "darwin" or platform.machine() != "arm64" or os.geteuid() == 0:
    parser.error("native acceptance requires an Apple Silicon Mac and a non-root user")
root = a.root.resolve()
try:
    validate_root(root, Path.home().resolve())
except ValueError as error:
    parser.error(str(error))
if root.exists():
    parser.error('test root must not exist')
source = a.brew_source.resolve()
upmacx = a.upmacx.resolve()
if not upmacx.is_file() or not os.access(upmacx, os.X_OK) or not (source / '.git').is_dir():
    parser.error('missing executable or Homebrew source repository')
root.mkdir(parents=True)
(root / 'tools').mkdir()
git_wrapper = root / 'tools/git'
git_wrapper.write_text('#!/bin/sh\nexport GIT_ALLOW_PROTOCOL=file\nexec /usr/bin/git "$@"\n')
git_wrapper.chmod(0o755)
for name in ['home', 'cache', 'logs', 'build-temp', 'apps', 'assets', 'evidence']:
    (root / name).mkdir()
brew = root / 'brew'
env = dict(os.environ)
# Isolation applies to this maintainer test only; the application preserves HOME.
for key in list(env):
    if key.startswith('HOMEBREW_'):
        del env[key]
env.update({
    'HOME': str(root / 'home'), 'HOMEBREW_CACHE': str(root / 'cache'),
    'HOMEBREW_LOGS': str(root / 'logs'), 'HOMEBREW_TEMP': str(root / 'build-temp'),
    'HOMEBREW_NO_AUTO_UPDATE': '1', 'HOMEBREW_NO_INSTALL_CLEANUP': '1',
    'HOMEBREW_NO_INSTALL_FROM_API': '1', 'HOMEBREW_NO_ANALYTICS': '1',
    'HOMEBREW_NO_ENV_HINTS': '1', 'HOMEBREW_NO_INSTALL_UPGRADE': '1',
    'HOMEBREW_CASK_OPTS': '--appdir=' + str(root / 'apps'),
    'GIT_ALLOW_PROTOCOL': 'file', 'HOMEBREW_GIT_PATH': str(git_wrapper),
    'PATH': str(brew / 'bin') + ':/usr/bin:/bin:/usr/sbin:/sbin', 'LC_ALL': 'C',
})
sequence = 0

def run(args, expected=0, input_text=None):
    global sequence
    sequence += 1
    result = subprocess.run([str(x) for x in args], env=env, cwd=root,
                            input=input_text, text=True, capture_output=True, timeout=180)
    # Evidence contains argv/outputs only, never an environment dump.
    record = {'argv': [str(x) for x in args], 'exit': result.returncode,
              'stdout': result.stdout, 'stderr': result.stderr}
    (root / 'evidence' / f'{sequence:03d}.json').write_text(json.dumps(record, indent=2))
    if result.returncode != expected:
        raise RuntimeError(f'{args}: exit {result.returncode}, expected {expected}\n{result.stdout}\n{result.stderr}')
    return result.stdout

def git(directory, *args):
    return run(['git', '-C', directory, *args])

def commit(directory, message):
    git(directory, 'add', '.')
    git(directory, '-c', 'user.name=upmacx acceptance fixture',
        '-c', 'user.email=acceptance@example.invalid', 'commit', '-qm', message)

# Freeze the Homebrew remote at the reviewed tag, so explicit brew update stays
# local and cannot upgrade Homebrew or fetch user taps during acceptance.
run(['git', 'init', '--bare', '-q', root / 'brew-remote.git'])
git(source, 'push', root / 'brew-remote.git', '7.0.10:refs/heads/main', '7.0.10:refs/tags/7.0.10')
git(root / 'brew-remote.git', 'symbolic-ref', 'HEAD', 'refs/heads/main')
run(['git', 'clone', '--branch', 'main', root / 'brew-remote.git', brew])
portable = source / 'Library/Homebrew/vendor/portable-ruby'
required = (brew / 'Library/Homebrew/vendor/portable-ruby-version').read_text().strip()
if not (portable / required / 'bin/ruby').is_file():
    raise RuntimeError('matching portable Ruby is absent; no automatic installation is allowed')
shutil.copytree(portable / required, brew / 'Library/Homebrew/vendor/portable-ruby' / required)
(brew / 'Library/Homebrew/vendor/portable-ruby/current').symlink_to(required)
assert run([brew / 'bin/brew', '--prefix']).strip() == str(brew)
# Empty local core/cask fixture taps prevent Homebrew's update-report from
# attempting to bootstrap an external tap. The Git wrapper additionally makes
# external protocols fail closed even when Homebrew filters subprocess env.
for core_name in ['core', 'cask']:
    local = root / (core_name + '-source')
    local.mkdir()
    (local / 'README.md').write_text('Local empty acceptance fixture; not the official tap.\n')
    run(['git', 'init', '-q', local])
    commit(local, 'Empty local fixture tap')
    remote = root / (core_name + '-remote.git')
    run(['git', 'init', '--bare', '-q', remote])
    git(local, 'push', '-q', remote, 'HEAD:refs/heads/main')
    git(remote, 'symbolic-ref', 'HEAD', 'refs/heads/main')
    destination = brew / 'Library/Taps/homebrew' / ('homebrew-' + core_name)
    destination.parent.mkdir(parents=True, exist_ok=True)
    run(['git', 'clone', '--branch', 'main', remote, destination])
version = run([brew / 'bin/brew', '--version'])
assert version.splitlines()[0] == 'Homebrew 7.0.10'

# All fixture repositories and package payloads are local and owned by this run.
tap_source = root / 'tap-source'
tap_source.mkdir()
(tap_source / 'Formula').mkdir()
(tap_source / 'Casks').mkdir()
run(['git', 'init', '-q', tap_source])
run(['git', 'init', '--bare', '-q', root / 'tap-remote.git'])
git(tap_source, 'remote', 'add', 'origin', root / 'tap-remote.git')
git(root / 'tap-remote.git', 'symbolic-ref', 'HEAD', 'refs/heads/main')

def package(name, version, cask=False, auto=False, broken=False):
    payload = f'#!/bin/sh\nprintf "{name} {version}\\n"\n'.encode()
    archive = root / 'assets' / f'{name}-{version}.tar.gz'
    with tarfile.open(archive, 'w:gz', format=tarfile.USTAR_FORMAT) as tf:
        entry = tarfile.TarInfo(name)
        entry.size = len(payload)
        entry.mode = 0o755
        tf.addfile(entry, io.BytesIO(payload))
    checksum = hashlib.sha256(archive.read_bytes()).hexdigest()
    if cask:
        preflight = 'preflight { odie "intentional fixture upgrade failure" }' if broken else ''
        text = f'''cask "{name}" do
  version "{version}"
  sha256 "{checksum}"
  url "{archive.as_uri()}"
  name "upmacx acceptance fixture"
  desc "Local disposable acceptance fixture"
  homepage "https://github.com/htooaunglynn/upmacx"
  {"auto_updates true" if auto else ""}
  {preflight}
  binary "{name}"
end
'''
        (tap_source / 'Casks' / f'{name}.rb').write_text(text)
    else:
        classname = ''.join(part.title() for part in name.split('-'))
        body = 'odie "intentional fixture upgrade failure"' if broken else f'bin.install "{name}"'
        text = f'''class {classname} < Formula
  desc "Local disposable acceptance fixture"
  homepage "https://github.com/htooaunglynn/upmacx"
  url "{archive.as_uri()}"
  version "{version}"
  sha256 "{checksum}"
  license "MIT"
  def install
    {body}
  end
end
'''
        (tap_source / 'Formula' / f'{name}.rb').write_text(text)

def publish_fixture(message):
    commit(tap_source, message)
    git(tap_source, 'push', '-q', 'origin', 'HEAD:refs/heads/main')

for name in ['fixture-tool', 'fixture-held', 'fixture-zbroken']:
    package(name, '1.0')
for name in ['fixture-cask', 'fixture-cask-held', 'fixture-auto', 'fixture-zbroken-cask']:
    package(name, '1.0', cask=True, auto=name == 'fixture-auto')
publish_fixture('Initial local test versions')
tap = brew / 'Library/Taps/upmacx/homebrew-acceptance'
tap.parent.mkdir(parents=True, exist_ok=True)
run(['git', 'clone', '--branch', 'main', root / 'tap-remote.git', tap])
for name in ([] if a.casks_only else ['fixture-tool', 'fixture-held', 'fixture-zbroken']):
    run([brew / 'bin/brew', 'install', '--formula', '--build-from-source', '--ignore-dependencies', 'upmacx/acceptance/' + name])
for name in ['fixture-cask', 'fixture-cask-held', 'fixture-auto', 'fixture-zbroken-cask']:
    run([brew / 'bin/brew', 'install', '--cask', 'upmacx/acceptance/' + name])
run([brew / 'bin/brew', 'pin', *['upmacx/acceptance/' + name for name in (['fixture-cask-held'] if a.casks_only else ['fixture-held', 'fixture-cask-held'])]])

# Actual zero-update schema and status before version changes.
empty = run([brew / 'bin/brew', 'outdated', '--json=v2'])
assert json.loads(empty) == {'formulae': [], 'casks': []}
(root / 'evidence' / 'empty.json').write_text(empty)
for name in ['fixture-tool', 'fixture-held']:
    package(name, '2.0')
for name in ['fixture-cask', 'fixture-cask-held', 'fixture-auto']:
    package(name, '2.0', cask=True, auto=name == 'fixture-auto')
publish_fixture('Outdated local test versions')

# Inherited greedy settings deliberately challenge the application's overrides.
env['HOMEBREW_UPGRADE_GREEDY'] = '1'
env['HOMEBREW_UPGRADE_GREEDY_CASKS'] = 'fixture-auto'
check = run([upmacx, '--check', '--providers', 'brew'])
assert '[brew] checked' in check and 'fixture-auto' not in check
if not a.casks_only:
    assert 'excluded fixture-held: pinned' in check or 'excluded upmacx/acceptance/fixture-held: pinned' in check
assert 'excluded fixture-cask-held: pinned' in check
# Acquire a native non-greedy candidate schema independently of CLI rendering.
env['HOMEBREW_UPGRADE_GREEDY'] = '0'
env['HOMEBREW_UPGRADE_GREEDY_CASKS'] = ''
raw_outdated = run([brew / 'bin/brew', 'outdated', '--json=v2'])
(root / 'evidence/outdated.json').write_text(raw_outdated)
assert 'fixture-auto' not in raw_outdated
env['HOMEBREW_UPGRADE_GREEDY'] = '1'
env['HOMEBREW_UPGRADE_GREEDY_CASKS'] = 'fixture-auto' 
for name in (['fixture-cask'] if a.casks_only else ['fixture-tool', 'fixture-cask']):
    assert run([brew / 'bin' / name]).strip() == name + ' 1.0'
# A nonterminal invocation must not consume consent or change a package.
run([upmacx, '--providers', 'brew'], expected=2)
for name in (['fixture-cask'] if a.casks_only else ['fixture-tool', 'fixture-cask']):
    assert run([brew / 'bin' / name]).strip() == name + ' 1.0'
result = run([upmacx, '--yes', '--providers', 'brew'])
assert '[brew] succeeded' in result
for name in (['fixture-cask'] if a.casks_only else ['fixture-tool', 'fixture-cask']):
    assert run([brew / 'bin' / name]).strip() == name + ' 2.0'
# Old formula keg proves upgrade didn't implicitly clean up the first version.
if not a.casks_only:
    assert (brew / 'Cellar/fixture-tool/1.0').exists()
assert run([brew / 'bin/fixture-auto']).strip() == 'fixture-auto 1.0'
if not a.casks_only:
    assert run([brew / 'bin/fixture-held']).strip() == 'fixture-held 1.0'
assert run([brew / 'bin/fixture-cask-held']).strip() == 'fixture-cask-held 1.0'
run([upmacx, '--yes', '--clean', '--providers', 'brew'])
if not a.casks_only:
    assert not (brew / 'Cellar/fixture-tool/1.0').exists()

# Real partial mutation: one approved upgrade succeeds before an intentional
# formula build failure; failed provider must retain the earlier old keg.
if a.casks_only:
    package('fixture-cask', '3.0', cask=True)
    package('fixture-zbroken-cask', '2.0', cask=True, broken=True)
else:
    package('fixture-tool', '3.0')
    package('fixture-zbroken', '2.0', broken=True)
publish_fixture('Controlled partial-failure test')
partial = run([upmacx, '--yes', '--clean', '--providers', 'brew'], expected=1)
assert '[brew] failed' in partial and 'completed:' in partial
assert 'running approved cleanup' not in partial
failure_record = json.loads((root / 'evidence' / f'{sequence:03d}.json').read_text())
assert 'intentional fixture upgrade failure' in failure_record['stdout'] + failure_record['stderr']
if a.casks_only:
    assert run([brew / 'bin/fixture-cask']).strip() == 'fixture-cask 3.0'
else:
    assert run([brew / 'bin/fixture-tool']).strip() == 'fixture-tool 3.0'
    assert (brew / 'Cellar/fixture-tool/2.0').exists()
print('PASS: real isolated Homebrew ' + ('cask' if a.casks_only else 'formula/cask') + ' discovery, pins, non-greedy casks, check-only, nonterminal refusal, upgrades, cleanup and partial failure')
print('Evidence:', root / 'evidence')
