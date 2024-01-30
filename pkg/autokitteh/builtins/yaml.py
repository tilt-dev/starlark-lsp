# starlib.yaml
# provides functions for working with yaml data


## Functions

def dumps(obj: any) -> str:
    """Serializes Starlark object to a yaml string


    Args:
      obj - input Starlark object


    Returns:
      yaml string


    Examples:
      encode to yaml
      ```
      load("encoding/yaml.star", "yaml")
      data = {"foo": "bar", "baz": True}
      res = yaml.dumps(data)
      ```
"""
    pass


def loads(src: str) -> any:
    """Reads a source yaml string to a Starlark object


    Args:
      src - input string of yaml data


    Returns:
      Starlark object


    Examples:
      load a yaml string
      ```
      load("encoding/yaml.star", "yaml")
      data = \"\"\"foo: bar
      baz: true
      \"\"\"
      d = yaml.loads(data)
      print(d)
      # Output: {"foo": "bar", "baz": True}
      ```
"""
    pass
