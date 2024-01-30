# starlib.json
# provides functions for working with json data


## Functions

def encode(obj: any) -> str:
    """Returns a JSON string representation of a Starlark data structure


    Args:
      obj - Starlark data structure


    Examples:
      encode a simple object as a JSON string
      ```
      load("encoding/json.star", "json")
      x = json.encode({"foo": ["bar", "baz"]})
      print(x)
      # Output: {"foo":["bar","baz"]}
      ```
"""
    pass


def decode(src: str) -> any:
    """Returns the Starlark representation of a string instance containing a JSON document. Decoding fails if src is not a valid JSON string.


    Args:
       src - string, must be valid JSON string


    Returns:
      Starlark data structure


    Examples:
      decode a JSON string into a Starlark structure
      ```
      load("encoding/json.star", "json")
      x = json.decode('{"foo": ["bar", "baz"]}')
      ```
"""
    pass


def indent(src: str, prefix: str="", indent: str="\t") -> str:
    """The indent function pretty-prints a valid JSON encoding, and returns a string containing the indented form. It accepts one required positional parameter, the JSON string, and two optional keyword-only string parameters, prefix and indent, that specify a prefix of each new line, and the unit of indentation.


    Args:
      src - source JSON string to encode
      prefix - optional string prefix that will be prepended to each line. default is ""
      indent - optional string that will be used to represent indentations. default is "\t"


    Examples:
      "pretty print" a valid JSON encoding
      ```
      load("encoding/json.star", "json")
      x = json.indent('{"foo": ["bar", "baz"]}')
      # print(x)
      # {
      #    "foo": [
      #      "bar",
      #      "baz"
      #    ]
      # }
      ```

      "pretty print" a valid JSON encoding, including optional prefix and indent parameters
      ```
      load("encoding/json.star", "json")
      x = json.indent('{"foo": ["bar", "baz"]}', prefix='....', indent="____")
      # print(x)
      # {
      # ....____"foo": [
      # ....________"bar",
      # ....________"baz"
      # ....____]
      # ....}
      ```
"""
