# mremotedec

[![Go Workflow](https://github.com/qrxnz/mremotedec/actions/workflows/go.yml/badge.svg)](https://github.com/qrxnz/mremotedec/actions/workflows/go.yml)

> A simple tool to decrypt mRemoteNG connection files (`.xml`).

mremotedec started as a side idea while solving a VulnLab machine. After fighting with clunky tooling for long enough, I decided to write my own — minimal, clean, and built the way I wanted it to work.

Big shoutout to [x4nt0n](https://x.com/x4nt0n?s=21) for solving the "Lock" machine together with me 💖

<img width="1186" height="752" alt="Image" src="https://github.com/user-attachments/assets/2f9a9cc5-eb60-44a3-afe9-3e0abc7ca85f" />

## 🧰 Features

- Supports **CBC** and **GCM** decryption modes.
- Handles **Full File Encryption**.
- Output in human-readable format or **CSV**.
- Written in Go for easy portability.

## 🛠️ Installation

### 📦 Binary Releases

Pre-compiled binaries for Linux, Windows, and macOS are available on the [Releases](https://github.com/qrxnz/mremotedec/releases) page.

### 🐹Using Go

You can install `mremotedec` directly using `go install`:

```bash
go install github.com/qrxnz/mremotedec@latest
```

### 🏗️ Build from Source

To build from source, you need to have [Go](https://go.dev/) installed.

```bash
git clone https://github.com/qrxnz/mremotedec.git
cd mremotedec
go build -o mremotedec .
```

Alternatively, if you have [Task](https://taskfile.dev/) installed, you can use:

```bash
task build
```

### ❄️ Using Nix

- **Run without installing**

```bash
nix run github:qrxnz/mremotedec
```

- **Add to a Nix Flake**

Add input in your flake like:

```nix
{
 inputs = {
   mremotedec = {
     url = "github:qrxnz/mremotedec";
     inputs.nixpkgs.follows = "nixpkgs";
   };
 };
}
```

With the input added you can reference it directly:

```nix
{ inputs, system, ... }:
{
  # NixOS
  environment.systemPackages = [ inputs.mremotedec.packages.${pkgs.system}.default ];
  # home-manager
  home.packages = [ inputs.mremotedec.packages.${pkgs.system}.default ];
}
```

- **Install imperatively**

```bash
nix profile install github:qrxnz/mremotedec
```

## 📖 Usage

### ⊹ ࣪ ˖ Basic Usage

Decrypt a standard `confCons.xml` file using the default mRemoteNG password (`mR3m`):

```bash
mremotedec confCons.xml
```

### 🔐 Custom Password

If you set a custom password in mRemoteNG, use the `-p` or `-password` flag:

```bash
mremotedec -p "your_custom_password" confCons.xml
```

### 🧾 CSV Output

Export the decrypted connections to a CSV file:

```bash
mremotedec -csv confCons.xml > connections.csv
```

## ❓ How it works

mRemoteNG uses different encryption methods based on its configuration:

- **Legacy (CBC):** Uses AES-CBC with an MD5-hashed password.
- **Modern (GCM):** Uses AES-GCM with PBKDF2 (SHA1) key derivation.

`mremotedec` automatically detects the encryption mode from the XML attributes (`BlockCipherMode`) and attempts to decrypt accordingly.

## 🗒️ Credits

### 🎨 Inspiration

I was inspired by:

- [gquere/mRemoteNG_password_decrypt](https://github.com/gquere/mRemoteNG_password_decrypt)

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
