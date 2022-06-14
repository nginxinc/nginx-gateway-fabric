import { default as hm } from '../src/httpmatches.js';

let expect = require('chai').expect;

// Creates a NGINX HTTP Request Object for testing.
// See documentation for all properties available: http://nginx.org/en/docs/njs/reference.html
function createRequest({ method = '', headers = {}, params = {}, matches = '' } = {}) {
  let r = {
    // Test mocks
    return(statusCode) {
      r.testReturned = statusCode;
    },
    internalRedirect(redirectPath) {
      r.testRedirectedTo = redirectPath;
    },
    error(msg) {
      console.log('\tngx_error:', msg);
    },
    variables: {},
  };

  if (method) {
    r.method = method;
  }

  if (headers) {
    r.headersIn = headers;
  }

  if (params) {
    r.args = params;
  }

  if (matches) {
    r.variables[hm.MATCHES_VARIABLE] = matches;
  }

  return r;
}

describe('extractMatchesFromRequest', () => {
  const tests = [
    {
      name: 'throws if matches variable does not exist on request',
      request: createRequest(),
      expectThrow: true,
      errSubstring: 'http_matches is not defined',
    },
    {
      name: 'throws if matches variable is not JSON',
      request: createRequest({ matches: 'not-JSON' }),
      expectThrow: true,
      errSubstring: 'error parsing',
    },
    {
      name: 'throws if matches variable is not an array',
      request: createRequest({ matches: '{}' }),
      expectThrow: true,
      errSubstring: 'expected a list of matches',
    },
    {
      name: 'throws if the length of the matches variable is zero',
      request: createRequest({ matches: '[]' }),
      expectThrow: true,
      errSubstring: 'matches is an empty list',
    },
    {
      name: 'does not throw if matches variable is expected list of matches',
      request: createRequest({ matches: '[{"any":true}]' }),
      expectThrow: false,
    },
  ];
  tests.forEach((test) => {
    it(test.name, () => {
      if (test.expectThrow) {
        expect(() => hm.extractMatchesFromRequest(test.request)).to.throw(test.errSubstring);
      } else {
        expect(() => hm.extractMatchesFromRequest(test.request).to.not.throw());
      }
    });
  });
});

describe('matchFound', () => {
  const tests = [
    {
      name: 'returns true if any is set to true',
      match: { any: true },
      request: createRequest(),
      expected: true,
    },
    {
      name: 'returns true if method matches and no other conditions are set',
      match: { method: 'GET' },
      request: createRequest({ method: 'GET' }),
      expected: true,
    },
    {
      name: 'returns true if headers match and no other conditions are set',
      match: { headers: ['header:value'] },
      request: createRequest({ headers: { header: 'value' } }),
      expected: true,
    },
    {
      name: 'returns true if query parameters match and no other conditions are set',
      match: { params: ['key=value'] },
      request: createRequest({ params: { key: 'value' } }),
      expected: true,
    },
    {
      name: 'returns true if multiple conditions match',
      match: { method: 'GET', headers: ['header:value'], params: ['key=value'] },
      request: createRequest({
        method: 'GET',
        headers: { header: 'value' },
        params: { key: 'value' },
      }),
      expected: true,
    },
    {
      name: 'returns false if method does not match',
      match: { method: 'POST' },
      request: createRequest({ method: 'GET' }),
      expected: false,
    },
    {
      name: 'returns false if headers do not match',
      match: { method: 'GET', headers: ['header:value'] },
      request: createRequest({ method: 'GET' }), // no headers are set on request
      expected: false,
    },
    {
      name: 'returns false if query parameters do not match',
      match: { method: 'GET', headers: ['header:value'], params: ['key=value'] },
      request: createRequest({ method: 'GET', headers: { header: 'value' } }), // no params set on request
      expected: false,
    },
    {
      name: 'throws if headers are malformed',
      match: { headers: ['malformedheader'] },
      request: createRequest(),
      expectThrow: true,
      errSubstring: 'invalid header match',
    },
    {
      name: 'throws if params are malformed',
      match: { params: ['keyvalue'] },
      request: createRequest(),
      expectThrow: true,
      errSubstring: 'invalid query parameter',
    },
  ];

  tests.forEach((test) => {
    it(test.name, () => {
      if (test.expectThrow) {
        expect(() => hm.matchFound(test.request, test.match)).to.throw(test.errSubstring);
      } else {
        const result = hm.matchFound(test.request, test.match);
        expect(result).to.equal(test.expected);
      }
    });
  });
});

describe('findMatch', () => {
  const headerMatch = { headers: ['header:value'] };
  const queryParamMatch = { params: ['key=value'] };
  const methodMatch = { method: 'POST' };
  const anyMatch = { any: true };
  const malformedMatch = { headers: ['malformed'] };

  const tests = [
    {
      name: 'returns first match that the request satisfies',
      matches: [headerMatch, queryParamMatch, methodMatch, anyMatch], // second match should be returned
      request: createRequest({ method: 'POST', params: { key: 'value' } }),
      expected: queryParamMatch,
    },
    {
      name: 'returns null when no match exists',
      matches: [headerMatch, queryParamMatch, methodMatch],
      request: createRequest({ method: 'GET' }),
      expected: null,
    },
    {
      name: 'throws if an exception occurs while finding a match',
      matches: [headerMatch, queryParamMatch, malformedMatch],
      request: createRequest({ method: 'GET' }),
      expectThrow: true,
      errSubstring: 'invalid header match',
    },
  ];

  tests.forEach((test) => {
    it(test.name, () => {
      test.request.variables = {
        http_matches: JSON.stringify(test.matches),
      };

      if (test.expectThrow) {
        expect(() => hm.findMatch(test.request, test.matches)).to.throw(test.errSubstring);
      } else {
        const result = hm.findMatch(test.request, test.matches);
        expect(result).to.equal(test.expected);
      }
    });
  });
});

describe('headersMatch', () => {
  const multipleHeaders = ['header1:VALUE1', 'header2:value2', 'header3:value3']; // case matters for header values

  const tests = [
    {
      name: 'throws an error if a header has multiple colons',
      headers: ['too:many:colons'],
      expectThrow: true,
    },
    {
      name: 'throws an error if a header has no colon',
      headers: ['wrong=delimiter'],
      requestHeaders: {},
      expectThrow: true,
    },
    {
      name: 'returns false if one of the header values does not match',
      headers: multipleHeaders,
      requestHeaders: {
        header1: 'VALUE1',
        header2: 'value2',
        header3: 'wrong-value', // this value does not match
      },
      expected: false,
    },
    {
      name: 'returns false if one of the header values case does not match',
      headers: multipleHeaders,
      requestHeaders: {
        header1: 'value1', // this value is not the correct case
        header2: 'value2',
        header3: 'value3',
      },
      expected: false,
    },
    {
      name: 'returns true if all headers match',
      headers: multipleHeaders,
      requestHeaders: {
        header1: 'VALUE1', // this value is not the correct case
        header2: 'value2',
        header3: 'value3',
      },
      expected: true,
    },
  ];

  tests.forEach((test) => {
    it(test.name, () => {
      if (test.expectThrow) {
        expect(() => hm.headersMatch(test.requestHeaders, test.headers)).to.throw(
          'invalid header match',
        );
      } else {
        expect(hm.headersMatch(test.requestHeaders, test.headers)).to.equal(test.expected);
      }
    });
  });
});

describe('paramsMatch', () => {
  const params = ['Arg1=value1', 'arg2=value2=SOME=other=value', 'arg3===value3&*1(*+']; // case matters for header values

  const tests = [
    {
      name: 'throws an error a param has no key',
      params: ['=nokey'],
      expectThrow: true,
    },
    {
      name: 'throws an error if a param has no value',
      params: ['novalue='],
      expectThrow: true,
    },
    {
      name: 'throws an error a param has no equal sign delimiter',
      params: ['keyval'],
      expectThrow: true,
    },
    {
      name: 'returns false if one of the params is missing from request',
      params: params,
      requestParams: ['arg2=value2=SOME=other=value', 'arg3===value3&*1(*+'],
      expected: false,
    },
    {
      name: 'returns false if one of the param values does not match',
      params: params,
      requestParams: ['Arg1=not-value-1', 'arg2=value2=SOME=other=value', 'arg3===value3&*1(*+'],
      expected: false,
    },
    {
      name: 'returns false if the case of one param values does not match',
      params: params,
      requestParams: ['Arg1=VALUE1', 'arg2=value2=SOME=other=value', 'arg3===value3&*1(*+'],
      expected: false,
    },
    {
      name: 'returns true if all params match',
      params: params,
      requestParams: params,
      expected: false,
    },
  ];

  tests.forEach((test) => {
    it(test.name, () => {
      if (test.expectThrow) {
        expect(() => hm.paramsMatch(test.requestParams, test.params)).to.throw(
          'invalid query parameter',
        );
      } else {
        expect(hm.paramsMatch(test.requestParams, test.params)).to.equal(test.expected);
      }
    });
  });
});

describe('redirect', () => {
  const testAnyMatch = { any: true, redirectPath: '/any' };
  const testHeaderMatches = {
    headers: ['header1:VALUE1', 'header2:value2', 'header3:value3'],
    redirectPath: '/headers',
  };
  const testQueryParamMatches = {
    params: ['Arg1=value1', 'arg2=value2=SOME=other=value', 'arg3===value3&*1(*+'],
    redirectPath: '/params',
  };
  const testAllMatchTypes = {
    method: 'GET',
    headers: ['header1:value1', 'header2:value2'],
    params: ['Arg1=value1', 'arg2=value2=SOME=other=value'],
    redirectPath: '/a-match',
  };

  const tests = [
    {
      name: 'returns Internal Server Error status code if http_matches variable is not set',
      request: createRequest(),
      matches: null,
      expectedReturn: hm.HTTP_CODES.internalServerError,
    },
    {
      name: 'returns Internal Server Error status code if http_matches contains malformed match',
      request: createRequest(),
      matches: [{ headers: ['malformedheader'] }],
      expectedReturn: hm.HTTP_CODES.internalServerError,
    },
    {
      name: 'returns Not Found status code if request does not satisfy any match',
      request: createRequest({ method: 'GET' }),
      matches: [{ method: 'POST' }],
      expectedReturn: hm.HTTP_CODES.notFound,
    },
    {
      name: 'returns Internal Server Error status code if request satisfies match, but the redirectPath is missing',
      request: createRequest({ method: 'GET' }),
      matches: [{ method: 'GET' }],
      expectedReturn: hm.HTTP_CODES.internalServerError,
    },
    {
      name: 'redirects to the redirectPath of the first match the request satisfies',
      request: createRequest({
        method: 'GET',
        headers: { header1: 'value1', header2: 'value2' },
        params: { Arg1: 'value1', arg2: 'value2=SOME=other=value' },
      }),
      matches: [testHeaderMatches, testQueryParamMatches, testAllMatchTypes, testAnyMatch], // request matches testAllMatchTypes and testAnyMatch. But first match should win.
      expectedRedirect: '/a-match',
    },
  ];

  tests.forEach((test) => {
    it(test.name, () => {
      if (test.matches) {
        // set http_matches variable
        test.request.variables = {
          http_matches: JSON.stringify(test.matches),
        };
      }

      hm.redirect(test.request);
      if (test.expectedReturn) {
        expect(test.request.testReturned).to.equal(test.expectedReturn);
      } else if (test.expectedRedirect) {
        expect(test.request.testRedirectedTo).to.equal(test.expectedRedirect);
      }
    });
  });
});
