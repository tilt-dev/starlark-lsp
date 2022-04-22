def any(x) -> bool:
  pass

def all(x) -> bool:
  pass

def bool(x) -> bool:
  pass

def bytes(x) -> Bytes:
  pass

def dict() -> Dict:
  pass

def dir(x) -> List[String]:
  pass

def enumerate(x) -> List[Tuple[int, any]]:
  pass

def float(x) -> float:
  pass

def hasattr(x, name) -> bool:
  pass

def hash(x) -> int:
  pass

def int(x) -> int:
  pass

def len(x) -> int:
  pass

def list() -> List:
  pass

def range() -> List[int]:
  pass

def repr(x) -> String:
  pass

def reversed(x) -> List:
  pass

def sorted(x) -> List:
  pass

def str(x) -> String:
  pass

def type(x) -> String:
  pass

def zip() -> List:
  pass

class Dict:
  def items(self) -> List:
    pass

  def keys(self) -> List:
    pass

  def update(self) -> None:
    pass

  def values(self) -> List:
    pass

class List:
  def append(self, x) -> None:
    pass

  def clear(self) -> None:
    pass

  def extend(self, x) -> None:
    pass

  def index(self, x) -> int:
    pass

  def insert(self, i, x) -> None:
    pass

  def remove(self, x) -> None:
    pass

class String:
  def capitalize(self) -> String:
    pass

  def count(self, sub) -> int:
    pass

  def endswith(self, suffix) -> bool:
    pass

  def find(self, sub) -> int:
    pass

  def format(self, *args, **kwargs) -> String:
    pass

  def index(self, sub) -> int:
    pass

  def isalnum(self) -> bool:
    pass

  def isalpha(self) -> bool:
    pass

  def isdigit(self) -> bool:
    pass

  def islower(self) -> bool:
    pass

  def isspace(self) -> bool:
    pass

  def istitle(self) -> bool:
    pass

  def isupper(self) -> bool:
    pass

  def join(self, iterable) -> String:
    pass

  def lower(self) -> String:
    pass

  def lstrip(self) -> String:
    pass

  def removeprefix(self, x) -> String:
    pass

  def removesuffix(self, x) -> String:
    pass

  def replace(self, old, new) -> String:
    pass

  def rfind(self, sub) -> int:
    pass

  def rindex(self, sub) -> int:
    pass

  def rsplit(self) -> List[String]:
    pass

  def rstrip(self) -> String:
    pass

  def split(self) -> List[String]:
    pass

  def splitlines(self) -> List[String]:
    pass

  def startswith(self, prefix) -> bool:
    pass

  def strip(self) -> String:
    pass

  def title(self) -> String:
    pass

  def upper(self) -> String:
    pass
