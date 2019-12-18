package main

const credentialsHowTo = `
Hello and welcome to sheets-refresh.

In order to start using it you need to create credentials file first.

To do that, please follow these steps:

https://github.com/rusq/xls2sheets#quickstart

1. Turn on the Google Sheets API described in Golang quickstart, and
   download the credentials.json file. If you need to tweak access,
   you can always do so in Google API & Services Console

2. Turn on the Google Drive API as described in drive quickstart. No
   need to dow nload credentials.json again, as it has already been
   downloaded on Step 1.

3. Copy or move it to: 
      %s
   and set mode 400 or 600 on the file if you're on osX or Linux.
   If you're using Microsoft Windows, don't worry.

4. Create a configuration file that will list the required source
   files and target spreadsheets (see Sample configuration).

5. During the first start you will be prompted to authorise
   application with your Google account.  There's no risk, as it is
   the application that was created on Step 1.  Once authorised, copy
   and paste the authorisation code from the browser into the prompt.

`
