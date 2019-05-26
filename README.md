# Excel To Google Sheets Importer #

Purpose: Import Microsoft Excel files from arbitrary location to
Google Sheets workbook.

### Quick install ###
If you have **Go** installed, run the following:
```sh
go get -u github.com/rusq/xls2sheets
go install github.com/rusq/xls2sheets/cmd/sheets-refresh
```
Otherwise, you can download the executable for your OS from [Releases][2]
page.

### Quickstart ###
1. Turn on the Google Sheets API described in Golang [quickstart][1], and
   download the `credentials.json` file.
2. Copy or move it to `$HOME/.refresh-credentials.json` and set mode 400 or
   600 on the file.
3. Create a configuration file that will list the required source files
   and target spreadsheets (see [Sample configuration](#example)).
4. During the first start you will be prompted to authorise application
   with your Google account.  Once authorised, copy and paste the
   authorisation code from the browser into the prompt.

### Configuration ###
* Configuration file describes a **Job** to be performed.
* A **Job** consists of one or more **Tasks**.
* Each **Task** has a name, and **Source** and **Target** sections.
  * In **Source** one must specify a *URI of the MS Excel file* (xlsx) and
    one or more *Address Ranges* to be processed, i.e. "*Workbook!A1:C1000*".
  * In **Target** - a *Google SpreadsheetID* and one or more *Address* to
    copy to, i.e. "Backup!A1".  Optionally, one can specify whether
    to Create or Clear the destination Sheet before copying.
  * It is important to have exactly same number of **Source Address Range**
    entries and **Target Addresses**.  I.e. if you're about to copy
    two sheets from an Excel file, make sure that you specify two target
    Google Spreadsheet Sheet addresses.

The Example file below contains all possible configuration entries.

#### Example ####
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
      - Data!A1:U
  target:
    spreadsheet_id: 1Qq9dCCj_DcnLE9lAOStEhhC37Crf7a77nBrKM-xhZZQ
    address:
      - Monthly Rates
    create: true
    clear: true
02_daily_rates:
  source:
    location: https://www.rbnz.govt.nz/-/media/ReserveBank/Files/Statistics/tables/b1/hb1-daily.xlsx
    address_range:
      - Data!A1:T
  target:
    spreadsheet_id: 1Qq9dCCj_DcnLE9lAOStEhhC37Crf7a77nBrKM-xhZZQ
    address:
      - Daily Rates
    create: true
    clear: true

```

### Sample Run ###
```
$ ./sheets-refresh -job rbrates.yaml

Go to the following link in your browser:
https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=XXXXXXXxxxxxxXXXXX.apps.googleusercontent.com&redirect_uri=urn%3Aietf%3Awg%3Aoauth%3A2.0%3Aoob&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fspreadsheets+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fdrive.file&state=state-token

Enter authorization code: 4/XxXxXxXxXxXxXxXx-ABCDEFG-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
Saving credential file to: /Users/you/Library/Caches/rusq/sheets-refresh/sheet-refresh-token.json
2019/05/19 18:37:19 Starting task: "01_monthly_rates"
2019/05/19 18:37:28 updating data in target spreadsheet 1Qq9dCCj_DcnLE9lAOStEhhC37Crf7a77nBrKM-xhZZQ
2019/05/19 18:37:28   * retrieving information about the spreadsheet
2019/05/19 18:37:29   * validating target configuration
2019/05/19 18:37:29   * copy range Data!A1:U to Monthly Rates
2019/05/19 18:37:30     * clearing target sheet
2019/05/19 18:37:31     * OK: 5209 cells updated
2019/05/19 18:37:32 Starting task: "02_daily_rates"
2019/05/19 18:37:37 updating data in target spreadsheet 1Qq9dCCj_DcnLE9lAOStEhhC37Crf7a77nBrKM-xhZZQ
2019/05/19 18:37:37   * retrieving information about the spreadsheet
2019/05/19 18:37:38   * validating target configuration
2019/05/19 18:37:38   * copy range Data!A1:T to Daily Rates
2019/05/19 18:37:40     * OK: 7061 cells updated
```

[1]: https://developers.google.com/sheets/api/quickstart/go
[2]: https://github.com/rusq/xls2sheets/releases