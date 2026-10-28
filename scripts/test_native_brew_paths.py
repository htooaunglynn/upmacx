import unittest
from pathlib import Path
from native_brew_paths import validate_root


class NativeLayoutTests(unittest.TestCase):
    def test_rejects_all_broad_write_locations(self):
        for location in ['/tmp/run', '/private/tmp/run', '/var/tmp/run',
                         '/private/var/tmp/run', '/private/var/folders/ab/cd/T/run',
                         '/private/var/folders/ab/cd/C/run', '/var/folders/ab/cd/T/run',
                         '/Users/test/Library/Developer/run',
                         '/Users/test/Library/Caches/org.swift.swiftpm/run']:
            with self.subTest(location=location):
                with self.assertRaises(ValueError):
                    validate_root(Path(location), Path('/Users/test'))

    def test_accepts_protected_project_prefix(self):
        validate_root(Path('/Users/test/Developer/upmacx/.native-tests/run'), Path('/Users/test'))

    def test_directory_boundary_not_string_prefix(self):
        validate_root(Path('/private/tmp-independent/run'), Path('/Users/test'))
        validate_root(Path('/Users/test/Library/Developer-independent/run'), Path('/Users/test'))


if __name__ == '__main__':
    unittest.main()
