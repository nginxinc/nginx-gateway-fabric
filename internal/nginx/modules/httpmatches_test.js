import redirect from './httpmatches.js';

let expect = require('chai').expect;

// NGINX HTTP Request Object.
//See documentation for all properties available: http://nginx.org/en/docs/njs/reference.html
let r = {
  // Test mocks
  return(statusCode) {
    r.testReturned = statusCode;
  },
  internalRedirect(redirectPath) {
    r.testRedirectedTo = redirectPath;
  },
  error(msg) {
    console.log('\terror:', msg);
  },
  warn(msg) {
    console.log('\twarn:', msg);
  },
};
const testHeaderMatches = {
  headers: ['header1:value1', 'header2:value2', 'header3:value3'],
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

describe('redirect', function () {
  beforeEach(function () {
    // reset fields on the test request object
    r.method = 'GET';
    r.variables = {};
    r.args = {};
    r.headersIn = {};
    // properties added for testing
    r.testReturned = 0;
    r.testRedirectedTo = '';
  });

  const tests = [
    {
      name: 'returns 500 if http_matches variable is not defined',
      expectedError: 500,
    },
    {
      name: 'returns 500 if http_matches is empty',
      matches: {},
      expectedError: 500,
    },
    {
      name: 'redirects to the redirectPath if no conditions are defined in match',
      matches: {
        redirectPath: 'no-conditions',
      },
      expectedPath: 'no-conditions',
    },
    {
      name: 'returns 405 if method does not match',
      matches: {
        method: 'GET',
        redirectPath: 'get-location',
      },
      expectedError: 405,
      requestModifier: (r) => {
        r.method = 'POST';
      },
    },
    {
      name: 'redirects to match redirectPath if method matches (GET)',
      matches: {
        method: 'GET',
        redirectPath: 'get-location',
      },
      expectedPath: 'get-location',
    },
    {
      name: 'redirects to match redirectPath if method matches (POST)',
      matches: {
        method: 'POST',
        redirectPath: 'post-location',
      },
      expectedPath: 'post-location',
      requestModifier: (r) => {
        r.method = 'POST';
      },
    },
    {
      name: 'returns 404 if no headers exist in request',
      matches: testHeaderMatches,
      expectedError: 404,
    },
    {
      name: 'returns 404 if not all headers exist in request',
      matches: testHeaderMatches,
      expectedError: 404,
      requestModifier: (r) => {
        r.headersIn = {
          header3: 'value3',
          header1: 'value1',
        };
      },
    },
    {
      name: 'returns 404 if all headers exist in request but not all match',
      matches: testHeaderMatches,
      expectedError: 404,
      requestModifier: (r) => {
        r.headersIn = {
          header3: 'value3',
          header1: 'value1',
          header2: 'not-value2',
        };
      },
    },
    {
      name: 'returns 500 if a header match is malformed',
      matches: { headers: ['header-without-a-colon'] },
      expectedError: 500,
    },
    {
      name: 'redirects to redirectPath if all headers exist in request and all match',
      matches: testHeaderMatches,
      expectedPath: '/headers',
      requestModifier: (r) => {
        r.headersIn = {
          header3: 'value3',
          header1: 'value1',
          header2: 'VALUE2', // header matching is case-insensitive. NGINX will lowercase the header name but not the value.
        };
      },
    },
    {
      name: 'returns 404 if no args exist in request',
      matches: testQueryParamMatches,
      expectedError: 404,
    },
    {
      name: 'returns 404 if not all args exist in request',
      matches: testQueryParamMatches,
      expectedError: 404,
      requestModifier: (r) => {
        r.args = {
          Arg1: 'value1',
          arg2: 'value2=SOME=other=value',
        };
      },
    },
    {
      name: 'returns 404 if all args exist in request but not all match',
      matches: testQueryParamMatches,
      expectedError: 404,
      requestModifier: (r) => {
        r.args = {
          Arg1: 'value1',
          arg2: 'value2=SOME=other=value',
          ARg3: '==value3&*1(*+', // query param matching is case sensitive, so this shouldn't match.
        };
      },
    },
    {
      name: 'returns 500 if param match is malformed (no "=")',
      matches: {
        params: ['arg-without-an-equal-sign'],
      },
      expectedError: 500,
    },
    {
      name: 'returns 500 if param match is malformed (no key)',
      matches: {
        params: ['=arg-without-a-key'],
      },
      expectedError: 500,
    },
    {
      name: 'returns 500 if param match is malformed (no value)',
      matches: {
        params: ['arg-without-a-value='],
      },
      expectedError: 500,
    },
    {
      name: 'redirects to redirectPath if all args exist in request and all match',
      matches: testQueryParamMatches,
      expectedPath: '/params',
      requestModifier: (r) => {
        r.args = {
          Arg1: 'value1',
          arg2: 'value2=SOME=other=value',
          arg3: '==value3&*1(*+',
        };
      },
    },
    {
      name: 'returns 405 if method does not match',
      matches: testAllMatchTypes,
      expectedError: 405,
      requestModifier: (r) => {
        r.method = 'POST';
        r.headersIn = {
          header1: 'value1',
          header2: 'value2',
        };
        r.args = {
          Arg1: 'value1',
          arg2: 'value2=SOME=other=value',
        };
      },
    },
    {
      name: 'returns 404 headers do not match',
      matches: testAllMatchTypes,
      expectedError: 404,
      requestModifier: (r) => {
        r.method = 'GET';
        r.headersIn = {
          header1: 'value1',
        };
        r.args = {
          Arg1: 'value1',
          arg2: 'value2=SOME=other=value',
        };
      },
    },
    {
      name: 'returns 404 if args do not match',
      matches: testAllMatchTypes,
      expectedError: 404,
      requestModifier: (r) => {
        r.method = 'GET';
        r.headersIn = {
          header1: 'value1',
          header2: 'value2',
        };
        r.args = {
          arg2: 'value2=SOME=other=value',
        };
      },
    },
    {
      name: 'redirects to redirectPath if all conditions match',
      matches: testAllMatchTypes,
      expectedPath: '/a-match',
      requestModifier: (r) => {
        r.method = 'GET';
        r.headersIn = {
          header1: 'value1',
          header2: 'value2',
        };
        r.args = {
          Arg1: 'value1',
          arg2: 'value2=SOME=other=value',
        };
      },
    },
  ];

  tests.forEach((test) => {
    it(test.name, () => {
      if (test.requestModifier) {
        test.requestModifier(r);
      }
      if (test.matches) {
        r.variables = {
          http_matches: JSON.stringify(test.matches),
        };
      }
      redirect.redirect(r);
      if (test.expectedPath) {
        expect(r.testRedirectedTo).to.equal(test.expectedPath);
      } else {
        expect(r.testReturned).to.equal(test.expectedError);
      }
    });
  });
});
