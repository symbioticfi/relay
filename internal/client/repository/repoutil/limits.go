package repoutil

// MaxSignatureRequestPageBytes bounds encoded values before decoding or copying.
// Pagination must constrain bytes as well as record count.
const MaxSignatureRequestPageBytes = 4 * 1024 * 1024
