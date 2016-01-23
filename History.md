
1.5.1 / 2016-01-23
==================

  * Fixed docker build
  * Fix: Do not try to load negative slice bounds

1.5.0 / 2015-07-10
==================

  * Display only 60 items in graph
  * Colorize current value label according to metric status

1.4.0 / 2015-07-07
==================

  * Added abbreviation to MAD
  * Write access-log
  * Migrated to Gorilla mux instead of martini
  * Added support for local file storage
  * Moved towards modular storage system
  * Added "value" parameter to API documentation and welcome runner

1.3.0 / 2015-07-06
==================

  * Added license file
  * Add statistical monitoring
  * Add "value" as an option and basic graphing capabilities

  Thanks [zainhoda](https://github.com/zainhoda) for the contribution containing graphing support!

1.2.2 / 2015-04-22
==================

  * Close the reader to ensure we don't spam with open FDs
  * Replaced URL to use apex

1.2.1 / 2015-02-20
==================

  * Building against go1.4.2
  * GOLINT: Reduced complexity of main function
  * GOLINT: Explicitly ignore errors
  * GOLINT: Unexported data types

1.2.0 / 2015-02-17
==================

  * Added Godeps to ensure build environment is stable
  * Added link to API docs to navbar
  * Added HTTPs protocol to API documentation

1.1.0 / 2015-02-08
==================

  * Do not show expired metrics
  * Fix: Store metricID in newly created metric
  * Some small bugfixes and imporovements

1.0.0 / 2015-02-07
==================

  * Initial version
