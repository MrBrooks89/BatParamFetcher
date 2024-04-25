**BatParamFetcher**

BatParamFetcher is a Go program that fetches URLs from the Wayback Machine for a list of domains, cleans the URLs by replacing query parameters with placeholders, and saves the cleaned URLs to files.
Installation

To install BatParamFetcher, you can use go install:

```
go install github.com/MrBrooks89/BatParamFetcher@latest
```
This will download and install the latest version of BatParamFetcher to your Go bin directory.
Usage

Use the following command to run BatParamFetcher:

BatParamFetcher -l path/to/domain-list.txt

Optional flags:
```
-h: Show usage information.
-l: Path to a file containing a list of domains (required).
-o: Path to the output directory for the cleaned URLs (default is results).
```
Example:

BatParamFetcher -l domains.txt -o output

BatParamFetcher will fetch URLs from the Wayback Machine for each domain, clean the URLs, and save the cleaned URLs to files in the specified output directory.
License

This project is licensed under the MIT License.
