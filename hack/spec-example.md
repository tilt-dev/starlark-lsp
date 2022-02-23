# Starlark Spec

blablabla

## Built-in constants and functions

blablabla

### None

`None` is the distinguished value of the type `NoneType`.

### True and False

`True` and `False` are the two values of type `bool`.

### any

`any(x)` returns `True` if any element of the iterable sequence x is true.
If the iterable is empty, it returns `False`.

### all

`all(x)` returns `False` if any element of the iterable sequence x is false.
If the iterable is empty, it returns `True`.

### bool

`bool(x)` interprets `x` as a Boolean value---`True` or `False`.
With no argument, `bool()` returns `False`.

### bytes

`bytes(x)` converts its argument to a `bytes`.

## Built-in methods

blablabla

<a id='bytes·elems'></a>
### bytes·elems

`b.elems()` returns an opaque iterable value containing successive int elements of b.
Its type is `"bytes.elems"`, and its string representation is of the form `b"...".elems()`.

```python
type(b"ABC".elems())	# "bytes.elems"
b"ABC".elems()	        # b"ABC".elems()
list(b"ABC".elems())  	# [65, 66, 67]
```
<!-- TODO: signpost how to convert an single int or list of int to a bytes. -->

<a id='dict·get'></a>
### dict·get

`D.get(key[, default])` returns the dictionary value corresponding to the given key.
If the dictionary contains no such value, `get` returns `None`, or the
value of the optional `default` parameter if present.

`get` fails if `key` is unhashable, or the dictionary is frozen or has active iterators.

```python
x = {"one": 1, "two": 2}
x.get("one")                            # 1
x.get("three")                          # None
x.get("three", 0)                       # 0
```

<a id='dict·items'></a>
### dict·items

`D.items()` returns a new list of key/value pairs, one per element in
dictionary D, in the same order as they would be returned by a `for` loop.

```python
x = {"one": 1, "two": 2}
x.items()                               # [("one", 1), ("two", 2)]
```

<a id='dict·keys'></a>
### dict·keys

`D.keys()` returns a new list containing the keys of dictionary D, in the
same order as they would be returned by a `for` loop.

```python
x = {"one": 1, "two": 2}
x.keys()                               # ["one", "two"]
```

## Grammar reference

blablabla
