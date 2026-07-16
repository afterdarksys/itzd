# Test-only DNSSEC zone

Every key in this directory is generated exclusively for hermetic tests of the
`resolver.test.` zone. It has no production authority or value. Never reuse
these keys, algorithms, tags, or DS records outside tests.

`zone.signed` contains secure, multiple-chain, and conflicting `_waddr` cases.
`cases.json` also identifies response states synthesized by transport tests,
including insecure, bogus, and NXDOMAIN.
