def sheets_a1_range(sheet_name: str|None, from: str|None, to: str|None):
    """https://developers.google.com/sheets/api/guides/concepts#expandable-1"""
    pass

def sheets_read_cell(spreadsheet_id: str, sheet_name: str|None, row_index: str, col_index: str, value_render_option: str|None):
    """Read a single cell"""
    pass

def sheets_read_range(spreadsheet_id: str, a1_range: str, value_render_option: str|None):
    """Read a range of cells"""
    pass

def sheets_set_background_color(spreadsheet_id: : str, a1_range: str, color: str):
    """Set the background color in a range of cells"""
    pass

def sheets_set_text_format(spreadsheet_id: str, a1_range: str, color: str|None, bold: str|None, italic: str|None, strikethrough: str|None, underline: str|None):
    """Set the text format in a range of cells"""
    pass

def sheets_write_cell(spreadsheet_id: str, sheet_name:str|None, row_index: str, col_index: str, value: str):
    """Write a single of cell"""
    pass

def sheets_write_range(spreadsheet_id: str, a1_range: str, data: str):
    """Write a range of cells"""
    pass
