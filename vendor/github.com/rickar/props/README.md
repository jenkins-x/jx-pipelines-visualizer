# props: Go (golang) library for handling Java-style property files

This library provides compatibility with Java property files for Go.

There are two main types provided:
* `Properties` - read and write property files in Java format
* `Expander` - replaces property references wrapped by '${}' at runtime (as 
found in Ant/Log4J/JSP EL/Spring)

The full Java property file format (including all comment types, line 
continuations, key-value separators, unicode escapes, etc.) is supported.

