# service4.1# service

## Policies
import can goes down and not up so 
siblings cannot import each other too

1. app
2. business
3. foundation
4. vendor

- foundation packages should not import logger, config, etc.
if you necessary to import, you should use a function that take a string as argument


- filename of the package in foundation package should be the same as the package name.

- one way to identify if the package is a containment package not a provide package is when it doesn't make sense to name the file same as package name like `utils` that may contain `marshaling` so such containment packages should not be placed in foundation. because they are not isolated single responsibility package maybe should place in business or application

- usually packages in foundation should have at most 4-5 files.
