package processor

// POSSIBILITIES:

// 1. see how many emails the same IP is sending out. if beyond a threshold then blacklist it
// 2. analyse the emails with a classifier (we need both good and bad emails)
// 3. follow the URLs and check if they deliver a malware
// 4. download the malware and analyse it
// 5. extract the attachment and analyse it
// 6. resolve the country of from the IP
