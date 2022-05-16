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
      return false;
    }
    let val = r.headersIn[kv[0]];
    if (!val || val.toLowerCase() !== kv[1]) {
      r.warn('header: ' + h + ' does not exist in request');
      return false;
    }
  }

  return true;
}

function paramsMatch(r, params) {
  for (let i = 0; i < params.length; i++) {
    let p = params[i];
    // get index of first = in string
    const idx = p.indexOf('=');
    // divide string into key value using the index
    let kv = [p.slice(0, idx), p.slice(idx + 1)];
    if (kv.length !== 2) {
      r.error('invalid query parameter: ' + p);
      return false;
    }
    const val = r.args[kv[0]];
    if (!val || val !== kv[1]) {
      r.warn('query parameter ' + p + ' not in request URI');
      return false;
    }
  }

  return true;
}

function redirect(r) {
  if (!r.variables[matchesVariable]) {
    r.error(
      'cannot redirect the request; the variable ' +
        matchesVariable +
        ' is not defined on the request object',
    );
    r.return(404);
    return;
  }

  const match = JSON.parse(r.variables[matchesVariable]);

  if (match.method && match.method !== r.method) {
    r.warn('request method: ' + r.method + ' does not match expected method: ' + match.method);
    r.return(405);
    return;
  }

  if (match.headers && !headersMatch(r, match.headers)) {
    r.return(404);
    return;
  }

  if (match.params && !paramsMatch(r, match.params)) {
    r.return(404);
    return;
  }

  // If we pass all the above checks then the request satisfies the http match conditions and we need to redirect to the path.
  // Make sure there is a path to redirect traffic to.
  if (!match.redirectPath) {
    r.error('no path defined in http match');
    r.return(404);
  }

  r.internalRedirect(match.redirectPath);
}
