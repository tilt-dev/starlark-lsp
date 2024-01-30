# starlib.html
# html defines jquery-like html selection & iteration functions for HTML documents


## Types

class HtmlSelection:
    "An HTML document for querying"
    pass

    def attr(name: str) str:
        """Gets the specified attribute's value for the first element in the Selection. To get the value for each element individually, use a looping construct such as each or map method

        Args:
          name - attribute name to get the value of
        """
        pass

    def children() -> HtmlSelection:
        """Gets the child elements of each element in the Selection
        """
        pass

    def children_filtered(selector: str) -> HtmlSelection:
        """Gets the child elements of each element in the Selection, filtered by the specified selector

        Args:
          selector - a query selector string to filter the current selection, returning a new selection
        """
        pass

    def contents(selector: str) -> HtmlSelection:
        """Gets the children of each element in the Selection, including text and comment nodes

        Args:
          a query selector string to filter the current selection, returning a new selection
        """
        pass

    def find(selector: str) HtmlSelection:
        """Gets the descendants of each element in the current set of matched elements, filtered by a selector

        Args:
          selector - a query selector string to filter the current selection, returning a new selection
        """
        pass

    def filter(selector: str) HtmlSelection:
        """Filter reduces the set of matched elements to those that match the selector string

        Args:
          selector - a query selector string to filter the current selection, returning a new selection
        """
        pass

    def get(i: int) -> HtmlSelection:
        """Retrieves the underlying node at the specified index. alias: eq

        Args:
          i - numerical index of node to get
        """
        pass

    def has(selector: str) -> HtmlSelection:
        """Reduces the set of matched elements to those that have a descendant that matches the selector

        Args:
          selector - a query selector string to filter the current selection, returning a new selection
        """
        pass

    def isSelector(selector: str) -> bool:
        """checks the current matched set of elements against a selector and returns true if at least one of these elements matches

        Args:
          a query selector string to filter the current selection, returning a new selection
        """
        pass

    def parent(selector: str) -> HtmlSelection:
        """Gets the parent of each element in the Selection

        Args:
          selector - a query selector string to filter the current selection, returning a new selection
        """
        pass

    def parents_until(selector: str) -> HtmlSelection:
        """Gets the ancestors of each element in the Selection, up to but not including the element matched by the selector

        Args:
          selector - a query selector string to filter the current selection, returning a new selection
        """
        pass

    def siblings() -> HtmlSelection:
        """Gets the siblings of each element in the Selection
        """
        pass

    def text() -> str:
        """Gets the combined text contents of each element in the set of matched elements, including descendants
        """
        pass

    def first(selector: str) -> HtmlSelection:
        """Gets the first element of the selection

        Args:
          selector - a query selector string to filter the current selection, returning a new selection
        """
        pass

    def last(selector: str) -> HtmlSelection:
        """Gets the last element of the selection

        Args:
          selector - a query selector string to filter the current selection, returning a new selection
        """
        pass

    def len() -> int:
        """Returns the number of the nodes in the selection
        """
        pass

    def eq(i: int) -> HtmlSelection:
        """Gets the element at index i of the selection

        Args:
          i - numerical index of node to get
        """
        pass

## Functions
def html(markup: str) -> HtmlSelection:
    """Parses an HTML document returning a selection at the root of the document

    Args:
      markup : string - html text to build a document from
    """
    pass
