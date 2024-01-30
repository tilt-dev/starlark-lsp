# starlib.geo
# defines geographic operations in two-dimensional space


## Types

class Point:
    "A two-dimensional point in space"
    x: float
    y: float
    lat: float  # alias
    lng: float  # alias

    def distance(self, p2: Point) -> float:
        """Euclidean Distance to the other point

        Args:
          p2 - point to measure distance to
        """
        pass

    def distanceGeodesic(self, p2: Point) -> float:
        """Distance on the surface of a sphere with the same radius as Earth

        Args:
          p2 - point to measure distance to
        """


class Line:
    "An ordered list of points that define a line"
    p1: Point
    p2: Point

    def length() -> float:
        "Euclidean Length"
        pass

    def lengthGeodesic() -> float:
        "Line length on the surface of a sphere with the same radius as Earth"
        pass


class Polygon:
    "An ordered list of closed lines (rings) that define a shape. lists of coordinates that wind clockwise are filled regions and  anti-clockwise represent holes"
    rings: list[Line]

class MultiPolygon:
    "MultiPolygon groups a list of polygons to behave like a single polygon"
    polygons: list[Polygon]


## Functions

def Point(x: float, y: float) -> Point:
    """Point constructor, takes an x(longitude) and y(latitude) value and returns a Point object


    Args:
      x : float - x-dimension value (longitude if using geodesic space) |
      y : float - y-dimension value (latitude if using geodesic space) |


    Examples:
      create a point at the Stonehenge prehistoric monument in the United Kingdom
      ```
      load("geo.star", "geo")
      # create a point at 51.1789° N, 1.8262° W, use negative y (latitude) value for west quadrant
      stonehenge = geo.Point(51.1789, -1.8262)
      print(stonehenge)
      # Output: (51.178900,-1.826200)
      ```
    """
    pass

def Line(points: list[ tuple[float, float] | Point] ) -> Line:
    """Line constructor. Takes either an array of coordinate pairs or an array of point objects and returns the line that connects them. Points do not need to be collinear, providing a single point returns a line with a length of 0

    Args:
      points - list of points on the line
"""
    pass

def Polygon(rings: list[Line | list[ tuple[float, float] | Point]] ) -> Polygon:
    """Polygon constructor. Takes a list of lists of coordinate pairs (or point objects) that define the outer boundary and any holes / inner boundaries that represent a polygon. In GIS tradition, lists of coordinates that wind clockwise are filled regions and  anti-clockwise represent holes.

    Args:
      rings - list of closed lines that constitute the polygon
    """
    pass

def MultiPolygon(polygons: list[Polygon]) -> MultiPolygon:
    """MultiPolygon constructor. MultiPolygon groups a list of polygons to behave like a single polygon

    Args:
    polygons - list of polygons
    """
    pass

def within(geom: Point|Line|Polygon, polygon: Polygon|MultiPolygon) -> bool:
    """Returns True if geom is entirely contained by polygon

    Args:
      geom : [point,line,polygon] - maybe-inner geometry
      polygon : [Polygon,MultiPolygon] - maybe-outer polygon
"""

def parseGeoJSON(data: str) -> tuple[geoms, properties]:
    """Parses string data in IETF-7946 (GeoJSON) format (https://tools.ietf.org/html/rfc7946) returning a list of geometries and equal-length list of properties for each geometry

    Args:
      data : string - string of GeoJSON data

    Examples:
      parse example
      ```
      load("geo.star", "geo")
      geo_json_string = \"\"\"
      {
      "type": "FeatureCollection",
      "features": [{
      "type": "Feature",
      "properties": {
      "name": "Coors Field"
      },
      "geometry": {
      "type": "Point",
      "coordinates": [-104.99404, 39.75621]
      }
      }, {
      "type": "Feature",
      "properties": {
      "name": "Busch Field"
      },
      "geometry": {
      "type": "Point",
      "coordinates": [-104.98404, 39.74621]
      }
      }]
      }
      \"\"\"
      (geoms, props) = geo.parseGeoJSON(geo_json_string)
      print(props)
      # Output: [{"name": "Coors Field"}, {"name": "Busch Field"}]
      ```
"""
