## 0.0.5 - 2021-03-27

- Added a --unique flag. This can also be used with --check to check that a
  file is both sorted and unique.
- Always close temp files before moving them. On Windows attempting to move an
  open file causes an error.

## 0.0.4 - 2020-12-29

- Fix handling of errors during initialization. These sorts of error could
  lead to a confusing panic instead of showing the actual error message.

- Handle the case where stdout is not connected to the terminal. Previously
  this caused an error during initialization.

- Replace file renaming with copying to handle the case where the temp file we
  sort into and the original file are not on the same partition.

- Fix bug where sorting wasn't stable in the presence of two
  case-insensitively identical lines (and possibly other similar scenarios).

## 0.0.3 - 2020-12-27

- Fix terminal width check. It was using the height as the width. In addition,
  it now makes the text width 90 characters if the terminal is wider than
  that.

## 0.0.2 - 2019-09-27

- The --check flag was not implemented and now it is.

## 0.0.1 - 2019-08-27

- First release upon an unsuspecting world.
