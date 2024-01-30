# starlib.base64
# defines base64 encoding & decoding functions, often used to represent binary as text.


## Functions

def encode(src: str, encoding: str = "standard") -> str:
    """Returns the base64 encoding of src


    Args:
      src - source string to encode to base64
      encoding - optional string to set encoding dialect. allowed values are: standard,standard_raw,url,url_raw


    Examples:
      encode a string as base64
      ```
      load("encoding/base64.star", "base64")
      encoded = base64.encode("hello world!")
      print(encoded)
      # Output: aGVsbG8gd29ybGQh
      ```"""
    pass

def decode(src: str, encoding: str = "standard") -> str:
    """Parses base64 input, giving back the plain string representation


    Args:
      src - source string of base64-encoded text
      encoding - optional string to set decoding dialect. allowed values are: standard, standard_raw, url, url_raw


    Examples:
      encode a string as base64
      ```
      load("encoding/base64.star", "base64")
      decoded = base64.decode("aGVsbG8gd29ybGQh")
      print(decoded)
      # Output: hello world!
      ```"""
    pass
