# mremotedec

A simple tool to decrypt mRemoteNG connection files (`.xml`).

## Features

- Supports **CBC** and **GCM** decryption modes.
- Handles **Full File Encryption**.
- Output in human-readable format or **CSV**.
- Written in Go for easy portability.

## Installation

```bash
go install github.com/qrxnz/mremotedec@latest
```

_(Note: Ensure your `go/bin` is in your PATH)_

## Usage

### Basic Usage

Decrypt a standard `confCons.xml` file using the default mRemoteNG password (`mR3m`):

```bash
mremotedec confCons.xml
```

### Custom Password

If you set a custom password in mRemoteNG, use the `-p` or `-password` flag:

```bash
mremotedec -p "your_custom_password" confCons.xml
```

### CSV Output

Export the decrypted connections to a CSV file:

```bash
mremotedec -csv confCons.xml > connections.csv
```

## How it works

mRemoteNG uses different encryption methods based on its configuration:

- **Legacy (CBC):** Uses AES-CBC with an MD5-hashed password.
- **Modern (GCM):** Uses AES-GCM with PBKDF2 (SHA1) key derivation.

`mremotedec` automatically detects the encryption mode from the XML attributes (`BlockCipherMode`) and attempts to decrypt accordingly.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
