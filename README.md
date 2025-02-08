# Clean and Copy
These scripts can quickly turn your hashcat potfile into a wordlist to use for further cracking. Additionally, it turns that pesky HEX output into readable cleartext.

## Usage
python3 clean_and_copy.py

# Match Hash To User
A script that quicky matches a user's NTLM hash from NTDS.DIT dump and finds it in the potfile. 

## Build
go build -o match-hash-to-user match-hash-to-user.go
## Usage
./match-hash-to-user <NTDS-DUMP HASHCAT-POTFILE OUTPUT-FILE> 
