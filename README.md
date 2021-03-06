
# Excel To Google Sheets Importer #

[![Build Status](https://travis-ci.com/rusq/xls2sheets.svg?branch=master)](https://travis-ci.com/rusq/xls2sheets)

Purpose: Import Microsoft Excel or Google Spreadsheet files from arbitrary
location to Google Sheets workbook.

Supported Sources:

  * Files types:
    * Google Sheets spreadsheet.
    * Microsoft Excel: **xls, xlsx**
    * Plain CSV: **csv** (NOTE: don't worry about specifying address range, it
      will be populated automatically for CSV files).
    * Open Office Spreadsheet: **ods**
    * Maybe others (supported by Google Sheets)
  * Local files or Remote (can fetch from URLs).

Supported Targets:

  * Google Sheets spreadsheet;
  * Save file to the local disk (all supported by Google Sheets formats, i.e.
    XLSX, XLS, ODS, PDF, TXT, CSV, HTML).

## Features ##

* Many-to-One: Multiple Source spreadsheets can be combined into one Google
  Sheets Document;
* One-to-Many: One source file can be split into several different Google
  Sheets Documents;
* Allows to specify the Range within the source to copy and a target
  worksheet, i.e. copy "Rates!A1:H20" from source to "Rates" worksheet in
  target GS document;
* Copy multiple worsheets (or ranges) to multiple target worksheets, i.e.:
  * Range "Rates!A1:H12" in source file to "Rates2019" worksheet in target;
  * Range "Rates!A13:H24" in source file to "Rates2020 worksheet in target;
* Exporting files to disk in a number of formats.

### Quick install ###
If you have **Go** installed, run the following:

```sh
go get -u github.com/rusq/xls2sheets
go install github.com/rusq/xls2sheets/cmd/sheets-refresh
```

Otherwise, you can download the executable for your Operating System from
[Releases][1] page.

### Quickstart ###
1. Turn on the Google Sheets API described in Golang [quickstart][2], and
   download the `credentials.json` file.  If you need to tweak access, you
   can always do so in [Google API & Services Console][3]
2. Turn on the Google Drive API as described in [drive quickstart][4].  No
   need to download `credentials.json` again, as it has already been
   downloaded on Step 1.
3. Copy or move it to `$HOME/.refresh-credentials.json` and set mode 400 or
   600 on the file.
4. Create a configuration file that will list the required source files and
   target spreadsheets (see [Sample configuration](#example)).
5. During the first start you will be prompted to authorise application with
   your Google account.  There's no risk, as it is the application that was
   created on Step 1.  Once authorised, copy and paste the authorisation
   code from the browser into the prompt.

### Configuration ###
* Configuration file describes a **Job** to be performed.
* A **Job** consists of one or more **Tasks**.
* Each **Task** has a name, and **Source** and **Target** sections.
  * In **Source** one must specify a *URI of the Spreadsheet file*  or ID
    of source Google Sheets Document and one or more *Address Ranges* to be
    processed, i.e. "*Workbook!A1:C1000*" or "*Sheet1!A2:U*".  No need to
    specify the address range for *CSV* file.
  * In **Target** - a *Google SpreadsheetID* and one or more *Address* to copy
    to, i.e. "Backup!A1".  Optionally, one can specify whether to *Create* the
    worksheet or *Clear* the destination worksheet before copying.
    Additionally, one can specify a filename for export in *Location*
    parameter (see example below).
  * It is important to have exactly same number of **Source Address Range**
    entries and **Target Addresses**.  I.e. if you're about to copy
    two sheets from an Excel file, make sure that you specify two target
    Google Spreadsheet Sheet addresses.

The Example file below contains all possible configuration entries.

#### Example ####

In the example two source files are combined into one Google Sheets Document:

* The range "Data!A1:U" of file *hb1-monthly.xlsx* is imported into "Monthly
  Rates" worksheet of Google Sheets Document

```yaml
# 
# Sample job for fetching RBNZ exchange sheets and load them into a
# test spreadsheet from https://www.rbnz.govt.nz/statistics/b1
#
# To use this file:
#   1. Create an empty Google Spreadsheet.
#   2. Copy and Paste the spreadsheet_id into this configuration file.
#   3. Compile and run sheets-refresh
#
# This should populate the empty spreadsheet with data from RBNZ website.
01_monthly_rates:
  source:
    location: https://www.rbnz.govt.nz/-/media/ReserveBank/Files/Statistics/tables/b1/hb1-monthly.xlsx
    address_range:
      - Data!A1:U   # address range for Data sheet.
      - Data        # complete import of Data sheet.
  target:
    spreadsheet_id: 1Qq9dCCj_DcnLE9lAOStEhhC37Crf7a77nBrKM-xhZZQ
    address:
      - Monthly Rates
      - Another Monthly Rates (full)
    create: true
    clear: true
  leave_junk: false     # leave temporary files.  May be used for debugging.
02_daily_rates:
  source:
    location: https://www.rbnz.govt.nz/-/media/ReserveBank/Files/Statistics/tables/b1/hb1-daily.xlsx
    address_range:
      - Data!A1:T
  target:
    spreadsheet_id: 1Qq9dCCj_DcnLE9lAOStEhhC37Crf7a77nBrKM-xhZZQ
    location: ./sample.ods    # save the file locally too.
    address:
      - Daily Rates
    create: true
    clear: true

```

### Sample Run ###
```
$ ./sheets-refresh -job rbrates.yaml
2019/12/09 20:07:56 callback server listening on localhost:6061
Please follow the Instructions in your browser to authorize sheets-refresh
or press [Ctrl]+[C] to cancel...
2019/12/09 20:08:07 Saving token file to: /Users/rustamgilyazov/Library/Caches/rusq/sheets-refresh/auth-token.bin
2019/12/09 20:08:07 starting task: "01_monthly_rates"
2019/12/09 20:08:07 + type detected as: remote file
2019/12/09 20:08:07 + opening: https://www.rbnz.govt.nz/-/media/ReserveBank/Files/Statistics/tables/b1/hb1-monthly.xlsx
2019/12/09 20:09:16 updating data in target spreadsheet 1Qq9dCCj_DcnLE9lAOStEhhC37Crf7a77nBrKM-xhZZQ
2019/12/09 20:09:16   * retrieving information about the spreadsheet
2019/12/09 20:09:17   * validating target configuration
2019/12/09 20:09:17   * copy range "Data!A1:U" to "Monthly Rates"
2019/12/09 20:09:18     * clearing target sheet
2019/12/09 20:09:20     * OK: 5356 cells updated
2019/12/09 20:09:20   * exporting to ./sample.ods
2019/12/09 20:09:27     * export OK
2019/12/09 20:09:28 task "01_monthly_rates": success
2019/12/09 20:09:28 starting task: "02_daily_rates"
2019/12/09 20:09:28 + type detected as: remote file
2019/12/09 20:09:28 + opening: https://www.rbnz.govt.nz/-/media/ReserveBank/Files/Statistics/tables/b1/hb1-daily.xlsx
2019/12/09 20:09:34 updating data in target spreadsheet 1Qq9dCCj_DcnLE9lAOStEhhC37Crf7a77nBrKM-xhZZQ
2019/12/09 20:09:34   * retrieving information about the spreadsheet
2019/12/09 20:09:34   * validating target configuration
2019/12/09 20:09:34   * copy range "Data!A1:T" to "Daily Rates"
2019/12/09 20:09:35     * clearing target sheet
2019/12/09 20:09:37     * OK: 9841 cells updated
2019/12/09 20:09:38 task "02_daily_rates": success
```

[1]: https://github.com/rusq/xls2sheets/releases
[2]: https://developers.google.com/sheets/api/quickstart/go
[3]: https://console.developers.google.com/apis/dashboard?authuser=0
[4]: https://developers.google.com/drive/api/v3/quickstart/go
[5]: https://developers.google.com/sheets/api/guides/concepts