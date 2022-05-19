export default { redirect };

const matchesVariable = 'http_matches';

// FIXME(osborn): Need to add special handling for repeated headers.
// Should follow guidance from https://www.rfc-editor.org/rfc/rfc7230.html#section-3.2.2.
function headersMatch(r, headers) {
  for (let i = 0; i < headers.length; i++) {
    const h = headers[i];
    const kv = h.split(':');
    if (kv.length !== 2) {
      r.error('invalid header match: ' + h);
      throw 500;
    }
    let val = r.headersIn[kv[0]];
    if (!val || val.toLowerCase() !== kv[1]) {
      r.warn('header: ' + h + ' does not exist in request');
      throw 404;
    }
  }
}

function paramsMatch(r, params) {
  for (let i = 0; i < params.length; i++) {
    let p = params[i];
    // get index of first = in string
    const idx = p.indexOf('=');
    // throw error if index is invalid
    // an index of -1 means "=" is not present
    // an index of 0 means there is no value
    // an index of length -1 means there is no key
    if (idx === -1 || (idx === 0) | (idx === p.length - 1)) {
      r.error('invalid query parameter: ' + p);
      throw 500;
    }

    // divide string into key value using the index
    let kv = [p.slice(0, idx), p.slice(idx + 1)];

    const val = r.args[kv[0]];
    if (!val || val !== kv[1]) {
      r.warn('query parameter ' + p + ' not in request URI');
      throw 404;
    }
  }
}

function redirect(r) {
  if (!r.variables[matchesVariable]) {
    r.error(
      'cannot redirect the request; the variable ' +
        matchesVariable +
        ' is not defined on the request object',
    );
    r.return(500);
    return;
  }

  const match = JSON.parse(r.variables[matchesVariable]);

  if (match.method && match.method !== r.method) {
    r.warn('request method: ' + r.method + ' does not match expected method: ' + match.method);
    r.return(405);
    return;
  }

  if (match.headers) {
    try {
      headersMatch(r, match.headers);
    } catch (e) {
      r.return(e);
      return;
    }
  }

  if (match.params) {
    try {
      paramsMatch(r, match.params);
    } catch (e) {
      r.return(e);
      return;
    }
  }

  // If we pass all the above checks then the request satisfies the http match conditions and we need to redirect to the path.
  // Make sure there is a path to redirect traffic to.
  if (!match.redirectPath) {
    r.error('no path defined in http match');
    r.return(500);
  }

  r.internalRedirect(match.redirectPath);
}
