# starlib.csv
# parses and writes comma-separated values files


## Functions

def read_all(source: str, comma: str = ",", comment: str = "",
             lazy_quotes : bool = False, trim_leading_space: bool = False,
             fields_per_record: int = 0, skip: int = 0) -> list[list[str]]:
    """Reads all rows from a source string, returning a list of string lists


    Args:
      source : str - input string of csv data
      comma : str - string to be used as a field delimiter, defaults to "," (a comma). comma must be a valid character and must not be \r, \n, or the Unicode replacement character (0xFFFD)
      comment : bool - comment string if provided or "", is the comment character. Lines beginning with the comment character without preceding whitespace are ignored. With leading whitespace the comment character becomes part of the field, even if trim_leading_space is True. comment must be a valid character and must not be \r, \n, or the Unicode replacement character (0xFFFD). It must also not be equal to comma.
      lazy_quotes : bool - If lazy_quotes is True, a quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field.
      trim_leading_space : bool - If trim_leading_space is True, leading white space in a field is ignored. This is done even if the field delimiter, comma, is white space.
      fields_per_record : int - fields_per_record is the number of expected fields per record. If fields_per_record is positive, read_all requires each record to have the given number of fields. If fields_per_record is 0, read_all sets it to the number of fields in the first record, so that future records must have the same field count. If fields_per_record is negative, no check is made and records may have a variable number of fields.
      skip : int - number of rows to skip, omitting from returned rows


    Examples:
      read a csv string into a list of string lists
      ```
      load("encoding/csv.star", "csv")
      data_str = \"\"\"type,name,number_of_legs
      dog,spot,4
      cat,spot,3
      spider,samantha,8
      \"\"\"
      data = csv.read_all(data_str)
      print(data)
      # Output: [["type", "name", "number_of_legs"], ["dog", "spot", "4"], ["cat", "spot", "3"], ["spider", "samantha", "8"]]
      ```
"""
    pass

def write_all(source: list[list[str]] ,comma: str =",") -> str:
    """Writes all rows from source to a csv-encoded string


    Args:
      source - array of arrays of strings to write to csv
      comma : string - comma is the field delimiter, defaults to "," (a comma). comma must be a valid character and must not be \r, \n, or the Unicode replacement character (0xFFFD). |

    Examples:
      write a list of string lists to a csv string
      ```
      load("encoding/csv.star", "csv")
      data = [
      ["type", "name", "number_of_legs"],
      ["dog", "spot", "4"],
      ["cat", "spot", "3"],
      ["spider", "samantha", "8"],
      ]
      csv_str = csv.write_all(data)
      ```
"""
    pass
