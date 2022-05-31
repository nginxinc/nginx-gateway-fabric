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
    // Header names are compared in a case-insensitive manner, meaning header name "FOO" is equivalent to "foo".
    // The NGINX request's headersIn object lookup is case-insensitive as well.
    // This means that r.headersIn['FOO'] is equivalent to r.headersIn['foo'].
    let val = r.headersIn[kv[0]];
    if (!val || val !== kv[1]) {
      throw 404;
    }
  }
}

function paramsMatch(r, params) {
  for (let i = 0; i < params.length; i++) {
    let p = params[i];
    // We store query parameter matches as strings with the format "key=value"; however, there may be more than one instance of "=" in the string.
    // To recover the key and value, we need to find the first occurrence of "=" in the string.
    const idx = p.indexOf('=');
    // Check for an improperly constructed query parameter match. There are three possible error cases:
    // (1) if the index is -1, then there are no "=" in the string (e.g. "keyvalue")
    // (2) if the index is 0, then there is no value in the string (e.g. "key=").
    // NOTE: While query parameter values are permitted to be empty, the Gateway API Spec forces the value to be a non-empty string.
    // https://github.com/kubernetes-sigs/gateway-api/blob/50e61865db9659111582080daa5ca1a91bbe265d/apis/v1alpha2/httproute_types.go#L375
    // (3) if the index is equal to length -1, then there is no key in the string (e.g. "=value").
    if (idx === -1 || (idx === 0) | (idx === p.length - 1)) {
      r.error('invalid query parameter: ' + p);
      throw 500;
    }

    // Divide string into key value using the index.
    let kv = [p.slice(0, idx), p.slice(idx + 1)];

    const val = r.args[kv[0]];
    if (!val || val !== kv[1]) {
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
    return;
  }

  r.internalRedirect(match.redirectPath);
}
