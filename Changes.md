## 0.0.4

* Fix handling of errors during initialization. These sorts of error could
  lead to a confusing panic instead of showing the actual error message.


## 0.0.3 - 2020-12-27

* Fix terminal width check. It was using the height as the width. In addition,
  it now makes the text width 90 characters if the terminal is wider than
  that.


## 0.0.2 - 2019-09-27

* The --check flag was not implemented and now it is.


## 0.0.1 - 2019-08-27

* First release upon an unsuspecting world.
