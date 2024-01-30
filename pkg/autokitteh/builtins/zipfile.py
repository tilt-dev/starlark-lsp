# starlib.zipfile
# reads & parses zip archives


## Functions

def ZipFile(data: str) -> ZipFile:
    """Returns an open zip archive for reading


    Args:
      data: string representation of a zipped archive

    Examples:
      download zip file and open
      ```
      load("zipfile.star", "ZipFile")
      load("http.star", "http")
      url = "http://testurl.org/sample.zip"
      raw = http.get(url).body()
      zf = ZipFile(raw)
      ```"""
    pass


## Types
class ZipFile:
    """a zip archive object"""
    def namelist() -> list:
        """return a list of files in the archive


        Args:


        Returns:


        Examples:
          get list of filenames from ZipFile

          ```
          load("zipfile.star", "ZipFile")
          zf = ZipFile(rawZipData)
          files = zf.namelist()
          print(files) # ["file1.txt", "file2.txt", etc ]
          ```
        """
        pass


    def open(filename: str) -> ZipInfo:
        """Opens a file for reading


        Args:
          filename: name of the file in the archive to open


        Returns:


        Examples:
          open file from ZipArchive as a ZipInfo

          ```
          load("zipfile.star", "ZipFile")
          zf = ZipFile(rawZipData)
          files = zf.namelist()
          filename = files[0]
          info = zf.open(filename) # can now use ZipInfo methods to read file
          ```
        """
        pass


class ZipInfo:
    "An information object for interacting with a Zip archive component"

    def read() -> str:
        """Reads the file, returning it's string representation


        Args:


        Returns:


        Examples:
          read file

          ```
          load("zipfile.star", "ZipFile")
          zf = ZipFile(rawZipData)
          info = zf.open("file1.txt")
          txt = info.read()
          print(txt) # prints the contents of the file
          ```"""
        pass
