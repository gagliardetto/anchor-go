# usage

```bash
go run --codec=borsh --debug --src=/path/to/idl.json
```

Code will be generated and saved to a timestamped folder inside `./generated/`.

---


sh -c "$(curl -sSfL https://release.solana.com/v1.7.4/install)"

Adding export PATH="$HOME/.local/share/solana/install/active_release/bin:$PATH" to $HOME/.profile
Adding export PATH="$HOME/.local/share/solana/install/active_release/bin:$PATH" to $HOME/.zprofile


Close and reopen your terminal to apply the PATH changes or run the following in your existing shell:
  export PATH="$HOME/.local/share/solana/install/active_release/bin:$PATH"


sudo apt-get update && sudo apt-get upgrade && sudo apt-get install -y pkg-config build-essential libudev-dev
cargo install --git https://github.com/project-serum/anchor --tag v0.11.1 anchor-cli --locked


cd $GOPATH/src/github.com/project-serum/anchor/examples
cd escrow
anchor build
cat target/idl/escrow.json


$GOPATH/src/github.com/project-serum/anchor/examples/escrow
$GOPATH/src/github.com/project-serum/anchor/examples/zero-copy
$GOPATH/src/github.com/project-serum/anchor/examples/typescript
$GOPATH/src/github.com/project-serum/anchor/examples/tutorial
$GOPATH/src/github.com/project-serum/anchor/examples/sysvars
$GOPATH/src/github.com/project-serum/anchor/examples/swap
$GOPATH/src/github.com/project-serum/anchor/examples/spl
$GOPATH/src/github.com/project-serum/anchor/examples/pyth
$GOPATH/src/github.com/project-serum/anchor/examples/multisig
$GOPATH/src/github.com/project-serum/anchor/examples/misc
$GOPATH/src/github.com/project-serum/anchor/examples/lockup
$GOPATH/src/github.com/project-serum/anchor/examples/interface
$GOPATH/src/github.com/project-serum/anchor/examples/ido-pool
$GOPATH/src/github.com/project-serum/anchor/examples/events
$GOPATH/src/github.com/project-serum/anchor/examples/errors
$GOPATH/src/github.com/project-serum/anchor/examples/composite
$GOPATH/src/github.com/project-serum/anchor/examples/chat
$GOPATH/src/github.com/project-serum/anchor/examples/cfo
$GOPATH/src/github.com/project-serum/anchor/examples/cashiers-check


cd $GOPATH/src/github.com/project-serum/anchor/examples/

  for d in ${1:=.}/*/ ; do (echo "\n:: Entering $(basename $d) ::" && cd "$d" && anchor build --idl $GOPATH/src/github.com/gagliardetto/anchor-go/idl_files/); done

  for d in ${1:=.}/*/ ; do (echo "\n:: Entering $(basename $d) ::" && cd "$d" && /bin/cp "target/idl/$(basename $d | tr "-" _).json" $GOPATH/src/github.com/gagliardetto/anchor-go/idl_files/); done



anchor idl parse -f program/src/lib.rs -o target/idl/basic_0.json
