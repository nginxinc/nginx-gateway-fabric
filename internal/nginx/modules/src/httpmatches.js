const MATCHES_VARIABLE = 'http_matches';
const HTTP_CODES = {
  notFound: 404,
  internalServerError: 500,
};

function redirect(r) {
  let matches;

  try {
    matches = extractMatchesFromRequest(r);
  } catch (e) {
    r.error(e.message);
    r.return(HTTP_CODES.internalServerError);
    return;
  }

  // Matches is a list of http matches in order of precedence.
  // We will accept the first match that the request satisfies.
  // If there's a match, redirect request to internal location block.
  // If an exception occurs, return 500.
  // If no matches are found, return 404.
  let match;
  try {
    match = findWinningMatch(r, matches);
  } catch (e) {
    r.error(e.message);
    r.return(HTTP_CODES.internalServerError);
    return;
  }

  if (!match) {
    r.return(HTTP_CODES.notFound);
    return;
  }

  if (!match.redirectPath) {
    r.error(
      `cannot redirect the request; the match ${JSON.stringify(
        match,
      )} does not have a redirectPath set`,
    );
    r.return(HTTP_CODES.internalServerError);
    return;
  }

  r.internalRedirect(match.redirectPath);
}

function extractMatchesFromRequest(r) {
  if (!r.variables[MATCHES_VARIABLE]) {
    throw Error(
      `cannot redirect the request; the variable ${MATCHES_VARIABLE} is not defined on the request object`,
    );
  }

  let matches;

  try {
    matches = JSON.parse(r.variables[MATCHES_VARIABLE]);
  } catch (e) {
    throw Error(
      `cannot redirect the request; error parsing ${r.variables[MATCHES_VARIABLE]} into a JSON object: ${e}`,
    );
  }

  if (!Array.isArray(matches)) {
    throw Error(`cannot redirect the request; expected a list of matches, got ${matches}`);
  }

  if (matches.length === 0) {
    throw Error(`cannot redirect the request; matches is an empty list`);
  }

  return matches;
}

function findWinningMatch(r, matches) {
  for (let i = 0; i < matches.length; i++) {
    try {
      let found = testMatch(r, matches[i]);
      if (found) {
        return matches[i];
      }
    } catch (e) {
      throw e;
    }
  }

  return null;
}

function testMatch(r, match) {
  // check for any
  if (match.any) {
    return true;
  }

  // check method
  if (match.method && r.method !== match.method) {
    return false;
  }

  // check headers
  if (match.headers) {
    try {
      let found = headersMatch(r.headersIn, match.headers);
      if (!found) {
        return false;
      }
    } catch (e) {
      throw e;
    }
  }

  // check params
  if (match.params) {
    try {
      let found = paramsMatch(r.args, match.params);
      if (!found) {
        return false;
      }
    } catch (e) {
      throw e;
    }
  }

  // all match conditions are satisfied so return true
  return true;
}

function headersMatch(requestHeaders, headers) {
  for (let i = 0; i < headers.length; i++) {
    const h = headers[i];
    const kv = h.split(':');

    if (kv.length !== 2) {
      throw Error(`invalid header match: ${h}`);
    }
    // Header names are compared in a case-insensitive manner, meaning header name "FOO" is equivalent to "foo".
    // The NGINX request's headersIn object lookup is case-insensitive as well.
    // This means that requestHeaders['FOO'] is equivalent to requestHeaders['foo'].
    let val = requestHeaders[kv[0]];

    if (!val) {
      return false;
    }

    // split on comma because nginx uses commas to delimit multiple header values
    const values = val.split(',');
    if (!values.includes(kv[1])) {
      return false;
    }
  }

  return true;
}

function paramsMatch(requestParams, params) {
  for (let i = 0; i < params.length; i++) {
    let p = params[i];
    // We store query parameter matches as strings with the format "key=value"; however, there may be more than one instance of "=" in the string.
    // To recover the key and value, we need to find the first occurrence of "=" in the string.
    const idx = params[i].indexOf('=');
    // Check for an improperly constructed query parameter match. There are three possible error cases:
    // (1) if the index is -1, then there are no "=" in the string (e.g. "keyvalue")
    // (2) if the index is 0, then there is no value in the string (e.g. "key=").
    // NOTE: While query parameter values are permitted to be empty, the Gateway API Spec forces the value to be a non-empty string.
    // https://github.com/kubernetes-sigs/gateway-api/blob/50e61865db9659111582080daa5ca1a91bbe265d/apis/v1alpha2/httproute_types.go#L375
    // (3) if the index is equal to length -1, then there is no key in the string (e.g. "=value").
    if (idx === -1 || (idx === 0) | (idx === p.length - 1)) {
      throw Error(`invalid query parameter: ${p}`);
    }

    // Divide string into key value using the index.
    let kv = [p.slice(0, idx), p.slice(idx + 1)];

    const val = requestParams[kv[0]];

    if (!val || val !== kv[1]) {
      return false;
    }
  }

  return true;
}

export default {
  redirect,
  testMatch,
  findWinningMatch,
  headersMatch,
  paramsMatch,
  extractMatchesFromRequest,
  HTTP_CODES,
  MATCHES_VARIABLE,
};
