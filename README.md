# AutoBackup

## Overview

The script will create a .zip archive of the specified directory including all subdirectories. If no destination is specified the script will use the source directory's name and will create the file in your current directory.

The script adds a timestamp to the filename automatically.

## Usage

You can run AutoBackup.exe with command prompt or PowerShell by specifying the source directory using -source flag. Make sure to use double quotes if there is a space in your path

Example `AutoBackup.exe --source "C:\Program Files\Go"`

You can specify your destination by using -destination flag: `AutoBackup.exe -source "C:\Program Files\Go" -destination C:\Users\Administrator\Desktop\goBackup.zip`

or

`AutoBackup.exe -source "C:\Program Files\Go" -destination C:\Users\Administrator\Desktop\` in which case the source directory name will be used.