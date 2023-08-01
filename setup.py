import os
import subprocess
import shlex
from wheel.bdist_wheel import bdist_wheel
from setuptools import setup, find_packages, Extension
from setuptools.command.build_ext import build_ext
from setuptools_scm import get_version


class bdist_wheel_universal(bdist_wheel):
    def get_tag(self):
        *_, plat = super().get_tag()
        return "py2.py3", "none", plat


def build_if_not_exist():
    if os.path.isfile("mdz/bin/mdz"):
        return
    version = get_version()
    print(f"build mdz from source ({version})")
    subprocess.call(["make", "mdz"])
    errno = subprocess.call(shlex.split(
        f"make build-release GIT_TAG={version}"
    ), cwd="mdz")
    assert errno == 0, f"mdz build failed with code {errno}"


class ModelzExtension(Extension):
    """A custom extension to define the OpenModelz extension."""


class ModelzBuildExt(build_ext):
    def build_extension(self, ext: Extension) -> None:
        if not isinstance(ext, ModelzExtension):
            return super().build_extension(ext)

        build_if_not_exist()


setup(
    name="openmodelz",
    use_scm_version=True,
    packages=find_packages("mdz"),
    include_package_data=True,
    data_files=[("bin", ["mdz/bin/mdz"])],
    zip_safe=False,
    ext_modules=[
        ModelzExtension(name="mdz", sources=["mdz/*"]),
    ],
    cmdclass=dict(
        build_ext=ModelzBuildExt,
        # sdist=SdistCommand,
        bdist_wheel=bdist_wheel_universal,
    ),
)
