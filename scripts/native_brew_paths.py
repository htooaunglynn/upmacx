"""Layout checks for real Homebrew acceptance; no subprocess or filesystem writes."""
import re
from pathlib import Path


def validate_root(root: Path, home: Path) -> None:
    """Reject prefixes inside Homebrew's broad sandbox write allowances.

    Homebrew's macOS Sandbox.allow_write_system_temp grants these directories
    after repository write denials. A checkout beneath one of them therefore
    fails the inherited-sandbox protection check, even in an ordinary terminal.
    The caller must pass canonical, absolute paths (Path.resolve()).
    """
    broad = [Path('/private/tmp'), Path('/tmp'), Path('/private/var/tmp'),
             Path('/var/tmp'), home / 'Library/Developer',
             home / 'Library/Caches/org.swift.swiftpm']
    canonical = root.as_posix()
    for base in broad:
        if root == base or base in root.parents:
            raise ValueError('test root must be outside system temporary/Xcode cache directories; '
                             'use "$PWD/.native-tests/<new-run>" so Homebrew can protect its executable')
    if re.match(r'^/(?:private/)?var/folders/[^/]+/[^/]+/[CT](?:/|$)', canonical):
        raise ValueError('test root is inside a macOS temporary/cache sandbox allowance; '
                         'use "$PWD/.native-tests/<new-run>"')
