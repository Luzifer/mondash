# 1.14.1 / 2018-09-16

  * Fix: Template dir missing in Image

# 1.14.0 / 2018-09-16

  * Update documentation to reflect staleness\_status
  * Update Dockerfile to reduce image size
  * Update repo-runner image
  * Add support for different stale status than Unknown

# 1.13.0 / 2018-04-08

  * Change progress bar behavior:  
    _Instead of just showing three bars with different colors which is quite useless overall now segments of the progress bar are generated according to the state of the check at the corresponding point of time. The effect is the viewer can see how recently the status was triggered instead just seeing a percentage._


# 1.12.0 / 2018-01-20

  * Add mondash-nagios wrapper
  * Update options in client

# 1.11.0 / 2017-11-24

  * Add option to hide value on dashboard

# 1.10.2 / 2017-11-24

  * Fix: Sorting was unstable

# 1.10.1 / 2017-11-22

  * Fix: Replace URL of my private page
  * Fix: MAD is no unit, put the abbr in front of the metric

# 1.10.0 / 2017-11-22

  * Use CDNJS CDN for libraries, update bootstrap and jquery
  * Meta: Remove old paragraph from README
  * Meta: Update Dockerfile

# 1.9.0 / 2017-11-22

  * Do not use MAD on example dashboard
  * Extract required filter from pongo2-addons
  * Switch to dep for vendoring, update libraries
  * Fix: Description for BaseURL was broken
  * Fix: Initialization of s3 object needed adjustment
  * Fix: Welcome dashboard contains no metrics for 1m
  * Lint: Remove not required conversion
  * Meta: Add automated asset building for Github
  * Meta: Replace license stub with full license text

# 1.8.1 / 2016-09-06

  * Fix: Use type Status instead of string
  * Fix: Add missing &#34;Value&#34; field
  * Fix: Set authorization token on client request

# 1.8.0 / 2016-09-05

  * Add client library to access MonDash

1.7.0 / 2016-03-27
==================

  * Added JSON version of dashboard
  * Fix: Metrics now can get stale

1.6.0 / 2016-03-27
==================

  * Allow hiding of MAD by flag
  * Include application version into response
  * Allow to use passed status instead of MAD

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
  * Added &amp;#34;value&amp;#34; parameter to API documentation and welcome runner

1.3.0 / 2015-07-06
==================

  * Added license file
  * Add statistical monitoring
  * Add &amp;#34;value&amp;#34; as an option and basic graphing capabilities

  Thanks [zainhoda](https://github.com/zainhoda) for the contribution containing graphing support!

1.2.2 / 2015-04-22
==================

  * Close the reader to ensure we don&amp;#39;t spam with open FDs
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