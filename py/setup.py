from setuptools import setup

setup(
    name='openlan-utils',
    version='4.0.1',
    author='Daniel Ding',
    author_email='danieldin95@163.com',
    packages=['openlan_utils'],
    entry_points={
        'console_scripts': [
            'openlan-utils = openlan_utils.__main__:main',
        ]
    },
    install_requires=[
        'requests==2.20.0',
        'ruamel.yaml==0.16.5'
    ],
)
