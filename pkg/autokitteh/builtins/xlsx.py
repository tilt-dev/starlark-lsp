# starlib.xlsx
# implements excel file readers in starlark. currently a highly-experimental package that will definitely change at some point in the future


## Types

class ExcelFile:
    "An excel file"

    def get_sheets() -> dict[str]:
        "Return a dict of sheet names in this excel file"
        pass

    def get_rows(sheetname) -> list[list[str]]:
        "Get all populated rows / columns as a list-of-list strings"
        pass


## Functions

def get_url(url: str) -> ExcelFile:
    "Fetch an excel file from a given url"
    pass
