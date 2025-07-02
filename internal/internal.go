package internal

func IsKnownBrokenIdlForTests(idlFilename string) bool {
	// These might be almost valid IDLs (maybe a missing address?)
	// but they need to be manually fixed.
	return isAnyOf(idlFilename,
		"2pbhpVLBKvqKXNbV6V2cvYME1dE9KCb39jZczsZnoTfu.json",
		"3rKpgHeTxgiretapm6F1Bv5cvric1JEi9ZAbfWa3H8LG.json",
		"effM4rzQbgZD8J5wkubJbSVxTgRFWtatQcQEgYuwqrR.json",
		"effRBsQPi2Exq4NWN6SPiCQk4E6BvXkqiBeu6saMxoi.json",
		"Gc35YGXTPUBYbLitZxXVi6cpJTFeyS8Y3LDrQB1fqvKL.json",
		"H3mSXnmYN3fChvRU6rhkLf9nGkytpEenSGm5DrgjFgHK.json",
		"JCiN3FoAn68Mx4JaQ546viikunXujsPNvoDFYdKupboM.json",
		"MGoV9M6YUsdhJzjzH9JMCW2tRe1LLxF1CjwqKC7DR1B.json",
		"mGovj4f5QhoRyhpEDFEpyX1h2CWvUfqJXz7431wshyu.json",
		"vmTE1MUq7EBnZrXTLRRn2W9G2UMG6MEuh6UHngs3DuQ.json",
	)
}

func isAnyOf(value string, values ...string) bool {
	for _, v := range values {
		if value == v {
			return true
		}
	}
	return false
}
