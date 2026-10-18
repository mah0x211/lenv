# LuaRocks signature verification fixture

luarocks-3.13.0.tar.gz and its detached .asc signature were downloaded unchanged
from https://luarocks.github.io/luarocks/releases/. They are public upstream
release artifacts used to test the actual downloaded publisher key, signature
verification, cache reuse and extraction. The archive includes LuaRocks'
license files.

luarocks-publisher.asc was downloaded from https://github.com/hishamhm.gpg.
It is a public-key fixture served over local HTTPS in tests. Production fetches
the current keys from that endpoint before each verification, without pinning
a fingerprint or embedding keys in the executable.

untrusted-signature.asc is a genuine detached signature over the literal
bytes `archive`, created with a disposable test key. The private key is not
retained. It tests rejection of signers other than the downloaded publisher.
