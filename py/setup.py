from setuptools import setup

setup(
    name='olutils',
    version='4.0.1',
    author='Daniel Ding',
    author_email='danieldin95@163.com',
    packages=['olutils'],
    entry_points={
        'console_scripts': [
            'openlan-utils = olutils.__main__:main',
        ]
    },
    install_requires=[
        'requests==2.20.0',
        'ruamel.yaml==0.16.5'
    ],
)
