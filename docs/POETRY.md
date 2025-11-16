# Poetry

Due to quirks with Poetry, Terrarium requires it to be installed during the build process.
By default, Terrarium will install `poetry<2.0.0` if it detects a `poetry.lock` file.

You can control this with the `--install-tool` flag.
You can also use it to install custom Poetry plugins.

> Note: any packages installed using this method are uninstalled when Poetry is no longer needed.
> If you require packages in the final build, add them to your `pyproject.toml`.

For example:

```shell
terrarium build --install-tool="poetry~=2.0.0 poetry-plugin-pypi-mirror"
```
