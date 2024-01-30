# starlib.math
# defines a Starlark module of additional mathematical functions. All functions accept both int and float values as arguments.


## Functions


def acos(x: int|float) -> float:
    """Return the arc cosine of x, in radians"""
    pass

def acosh(x: int|float) -> float:
    """Return the inverse hyperbolic cosine of x"""
    pass

def asin(x: int|float) -> float:
    """Return the arc sine of x, in radians"""
    pass

def asinh(x: int|float) -> float:
    """Return the inverse hyperbolic sine of x"""
    pass

def atan(x: int|float) -> float:
    """Return the arc tangent of x, in radians"""
    pass

def atan2(y: int|float, x: int|float) -> float:
    """Return atan(y / x), in radians. The result is between -pi and pi. The vector in the plane from the origin to point (x, y) makes this angle with the positive X axis. The atan2() function can compute the correct quadrant for the angle since it knows the sign of both inputs. For example, atan(1) and atan2(1, 1) are both pi/4, but atan2(-1, -1) is -3*pi/4"""
    pass

def atanh(x: int|float) -> float:
    """Return the inverse hyperbolic tangent of x"""
    pass

def ceil(x: int|float) -> float:
    """Return the ceiling of x, the smallest integer greater than or equal to x"""
    pass

def copysign(x,y)
    """Returns a value with the magnitude of x and the sign of y"""
    pass

def cos(x: int|float) -> float:
    """Return the cosine of x radians"""
    pass

def cosh(x: int|float) -> float:
    """Return the hyperbolic cosine of x"""
    pass

def degrees(x: int|float) -> float:
    """Convert angle x from radians to degrees"""
    pass

def exp(x: int|float) -> float:
    """Returns e raised to the power x, where e = 2.718281… is the base of natural logarithms"""
    pass

def fabs(x: int|float) -> float:
    """Return the absolute value of x"""
    pass

def floor(x: int|float) -> float:
    """Return the floor of x, the largest integer less than or equal to x

    Examples:
      calculate the floor of 2.9
      ```
      load("math.star", "math")
      x = math.floor(2.9)
      print(x)
      # Output: 2
      ```
    """
pass

def gamma(x: int|float) -> float:
    """Returns the Gamma function of x"""
    pass

def hypot(x: int|float, y: int|float)
    """Return the Euclidean norm, sqrt(x*x + y*y). This is the length of the vector from the origin to point (x, y)"""
    pass

def log(x: int|float, base: int) -> float:
    """Returns the logarithm of x in the given base, or natural logarithm by default"""
    pass

def mod(x: int|float, y: int|float) -> float:
    """Returns the floating-point remainder of x/y. The magnitude of the result is less than y and its sign agrees with that of x"""
    pass

def pow(x: int|float, y: int|float) -> float:
    """Returns x**y, the base-x exponential of y

    Examples:
      raise 4 to the power of 3
      ```
      load("math.star", "math")
      x = math.pow(4,5)
      print(x)
      # Output: 1024.0
      ```
    """
    pass

def radians(x: int|float) -> float:
    """Convert angle x from degrees to radians"""
    pass

def remainder(x: int|float, y: int|float)
    """Returns the IEEE 754 floating-point remainder of x/y"""
    pass

def round(x: int|float) -> float:
    """Returns the nearest integer, rounding half away from zero"""
    pass

def sqrt(x: int|float) -> float:
    """Return the square root of x"""
    pass

def sin(x: int|float) -> float:
    """Return the sine of x radians"""
    pass

def sinh(x: int|float) -> float:
    """Return the hyperbolic sine of x"""
    pass

def tan(x: int|float) -> float:
    """Return the tangent of x radians"""
    pass

def tanh(x: int|float) -> float:
    """Return the hyperbolic tangent of x"""
    pass
