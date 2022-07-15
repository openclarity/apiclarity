# Trace Analyzer Module

This module aims at detecting insecure practices by analyzing API traces.

Each trace is analysed by a set of analyzers, each analyzer is specialized in a
particular kind of findings.

The analyzers are:
* WeakBasicAuth
* WeakJWT
* Sensitive information
* Guessable ID
* NLID

Those findings can be presented either at the API level or at the event level
depending on their type. Moreover findings at the API level can be deleted if
the user thinks they are not relevant anymore. Because the problem was fixed for
example.

## Analyzers
### Weak Basic Authentication

If the Basic Auth method is used:
    - check password length
    - check if the password is in a list of known weak passwords
    - check if the same username/password combination is used in multiple APIs

### Weak JSON Web Tokens

If a JWT is used:
    - check if no algorithm is set
    - check if signing algorithm is set to none
    - check for recommended signing algorithm
    - check for sensitive data in the token claims
    - check if there is a claim to expire the token
    - attempt a dictionary attack on the secret signing key. Dictionary can be provided in configuration

### Sensitive information

Apply a set of regexps rules to the headers and body of the request and the
response and raise alarms accordingly.
Those rules are stored in one or multiple files configured with the
`TRACE_ANALYZER_RULES_FILENAMES` configuration.

The format of a rules file is a yaml list.
Each element of the list is such that:

```yaml
- id: core-001
  description: Find 'password' keyword in flow data
  regex: '([pP][aA][sS][sS][wW][oO][rR][dD])'
  searchIn:   # Allowed values: RequestBody, ResponseBody, RequestHeaders, ResponseHeaders
    - RequestBody
    - ResponseBody
    - RequestHeaders
    - ResponseHeaders
```

The `regexp` field is a regular expression compatible with the RE2 format. See
https://github.com/google/re2/wiki/Syntax for more information.

### Guessable ID

This analyzer aims at finding identifiers that seem guessable.

For example, if an attacker sees that identifiers for a kind of object is
"0001", "0002", "0003" ... they can easily try to guess identifiers.

This analyzer raise a warning when such identifiers are detected.

### NLID: Non learnt identifier

Non learnt identifier detection can help detecting Broken Object Level
Authorization (BOLA) attacks.

It raises an alert when there is an attempt to manipulate a resource by its
identifier without having retrieved the identifier first.

## Configuration

Default dictionaries and rules are provided as part of the module (see
https://github.com/openclarity/apiclarity/tree/master/backend/pkg/modules/assets/traceanalyzer)

You can overwrite those files with the following environment variables:

TRACE_ANALYZER_DICT_FILENAMES
: Colon separated list of filenames containing list of known passwords.

`TRACE_ANALYZER_DICT_FILENAMES="/opt/custom/dictfile1.txt:/opt/custom/dictfile2.txt"`

TRACE_ANALYZER_RULES_FILENAMES
: Colon separated list of filenames containing rules for the `sensitive`
  analyzer.

TRACE_ANALYZER_SENSITIVE_KEYWORDS_FILENAMES
: Colon separated list of filenames containing keywords that can be considered
  as sensitive. For example, it's used by the Weak JWT analyzer in order to
  check for sensitive claim names.

TRACE_ANALYZER_IGNORE_FINDINGS
: Comma separated list of findings that must be ignored.
`TRACE_ANALYZER_IGNORE_FINDINGS=JWT_SENSITIVE_CONTENT_IN_CLAIMS,JWT_WEAK_SYMETRIC_SECRET`

## Credits

Example dictionnary files of known password are part of https://github.com/danielmiessler/SecLists
