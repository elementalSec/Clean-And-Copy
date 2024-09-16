#!/usr/bin/env python3

import os
from time import sleep
from pathlib import Path
import shutil
import binascii


def decode_hex(hex_value):
    try:
        encodings = [
            'ascii',
            'big5',
            'big5hkscs',
            'cp037',
            'cp273',
            'cp424',
            'cp437',
            'cp500',
            'cp720',
            'cp737',
            'cp775',
            'cp850',
            'cp852',
            'cp855',
            'cp856',
            'cp857',
            'cp858',
            'cp860',
            'cp861',
            'cp862',
            'cp863',
            'cp864',
            'cp865',
            'cp866',
            'cp869',
            'cp874',
            'cp875',
            'cp932',
            'cp949',
            'cp950',
            'cp1006',
            'cp1026',
            'cp1125',
            'cp1140',
            'cp1250',
            'cp1251',
            'cp1252',
            'cp1253',
            'cp1254',
            'cp1255',
            'cp1256',
            'cp1257',
            'cp1258',
            'euc_jp',
            'euc_jis_2004',
            'euc_jisx0213',
            'euc_kr',
            'gb2312',
            'gbk',
            'gb18030',
            'hz',
            'iso2022_jp',
            'iso2022_jp_1',
            'iso2022_jp_2',
            'iso2022_jp_2004',
            'iso2022_jp_3',
            'iso2022_jp_ext',
            'iso2022_kr',
            'latin_1',
            'iso8859_2',
            'iso8859_3',
            'iso8859_4',
            'iso8859_5',
            'iso8859_6',
            'iso8859_7',
            'iso8859_8',
            'iso8859_9',
            'iso8859_10',
            'iso8859_13',
            'iso8859_14',
            'iso8859_15',
            'iso8859_16',
            'johab',
            'koi8_r',
            'koi8_t',
            'koi8_u',
            'kz1048',
            'mac_cyrillic',
            'mac_greek',
            'mac_iceland',
            'mac_latin2',
            'mac_roman',
            'mac_turkish',
            'ptcp154',
            'shift_jis',
            'shift_jis_2004',
            'shift_jisx0213',
            'utf_32',
            'utf_32_be',
            'utf_32_le',
            'utf_16',
            'utf_16_be',
            'utf_16_le',
            'utf_7',
            'utf_8',
            'utf_8_sig',
        ]

        decoded_values = {}
        for encoding in encodings:
            try:
                decoded_value = binascii.unhexlify(hex_value).decode(encoding)
                decoded_values[encoding] = decoded_value
            except UnicodeDecodeError:
                continue
        return decoded_values
    except binascii.Error:
        return None

def parent_path_exists(dest):
    path=Path(dest)
    if path.parent.exists():
        return ""
    else:
        return str(path.parent)

home = str(Path.home())

# Take input for source and destination paths of the potfile
#source_path = input("Where is your hashcat potfile? (Full path with filename): ")
source_path = os.path.join(home, ".local/share/hashcat/hashcat.potfile")
destination_path = input("Where do you want to place the potfile? (Full path with filename): ")
parent_path = parent_path_exists(destination_path)


#Copy potfile from source to destination
#os.system(f"cp {source_path} {destination_path}")
if parent_path == "":
    shutil.copy(source_path, destination_path)
    
    sleep(0.5)
    file_name = "cracked-passes"
    #Check to see if cracked-passes exists and if it does, remove it before creating the new one
    if os.path.isfile(file_name):
        try:
            os.remove(file_name)
            print(f"Removed {file_name}")
        except Exception as e:
            print(f"an error occurred while trying to remove {file_name}: {e}")
    else:
        print(f"{file_name} does not exist in the current directory")

#Clean up potfile
#os.system(f"cat {destination_path} | cut -d ':' -f 2 >> cracked-passes")
    with open("cracked-passes", 'w') as f:
        lines_in_dest = open(destination_path, 'r').readlines()
        for line in lines_in_dest:
          if len(line) > 0:
            last_colon = line.rfind(':')
            password = line[last_colon + 1:].strip()
            if "HEX" in password:
                decoded_values = decode_hex(password)

                if decoded_values:
                    if decoded_values.get('ascii') is not None:
                        print(f"ascii: {decoded_values['ascii']}")
                        f.write(f"{hash}:{decoded_values['ascii']}\n")
                    elif decoded_values.get('utf_8') is not None:
                        print(f"utf_8: {decoded_values['utf_8']}")
                        f.write(f"{hash}:{decoded_values['utf_8']}\n")
                    elif decoded_values.get('big5') is not None:
                        print(f"big5: {decoded_values['big5']}")
                        f.write(f"{hash}:{decoded_values['big5']}\n")
                    else:
                        for encoding, value in decoded_values.items():
                            print(f"{encoding}: {value}")
            f.write(f"{password}\n")


    print("Completed")

else:
    print(f"Parent path does not exist {parent_path} for {destination_path}")
