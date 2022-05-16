import redirect from './httpmatches.js';
let expect = require('chai').expect;

// NGINX HTTP Request Object.
//See documentation for all properties available: http://nginx.org/en/docs/njs/reference.html
let r;

describe('redirect', function () {
  beforeEach(function () {
    r = {
      method: 'GET',
      variables: {},
      args: {},
      headersIn: {},
      // properties added for testing
      testReturned: 0,
      testRedirectedTo: '',
    };

    // Test mocks
    r.return = function (statusCode) {
      r.testReturned = statusCode;
    };

    r.internalRedirect = function (redirectPath) {
      r.testRedirectedTo = redirectPath;
    };

    r.error = function (msg) {
      console.log('\terror:', msg);
    };

    r.warn = function (msg) {
      console.log('\twarn:', msg);
    };
  });
  it('returns 404 if http_matches variable is not defined', function () {
    redirect.redirect(r);
    expect(r.testReturned).to.equal(404);
  });
  it('returns 404 if http_matches is empty', function () {
    r.variables = {
      http_matches: '{}',
    };
    redirect.redirect(r);
    expect(r.testReturned).to.equal(404);
  });
  it('redirects to the redirectPath if no conditions are defined in match', function () {
    let matches = {
      redirectPath: 'no-conditions',
    };
    r.variables = {
      http_matches: JSON.stringify(matches),
    };

    redirect.redirect(r);
    expect(r.testRedirectedTo).to.equal('no-conditions');
  });
  it('returns 405 if method does not match', function () {
    let matches = {
      method: 'GET',
      redirectPath: 'get-location',
    };
    r.method = 'POST';
    r.variables = {
      http_matches: JSON.stringify(matches),
    };
    redirect.redirect(r);
    expect(r.testReturned).to.equal(405);
  });
  it('redirects to match redirectPath if method matches', function () {
    let matches = {
      method: 'GET',
      redirectPath: 'get-location',
    };

    r.variables = {
      http_matches: JSON.stringify(matches),
    };
    redirect.redirect(r);
    expect(r.testRedirectedTo).to.equal('get-location');
    matches.method = 'POST';
    matches.redirectPath = 'post-location';
    r.method = 'POST';
    r.variables.http_matches = JSON.stringify(matches);
    redirect.redirect(r);
    expect(r.testRedirectedTo).to.equal('post-location');
  });
  describe('header matching', function () {
    beforeEach(function () {
      let matches = {
        headers: ['header1:value1', 'header2:value2', 'header3:value3'],
        redirectPath: '/headers',
      };
      r.variables = {
        http_matches: JSON.stringify(matches),
      };
    });
    it('returns 404 if no headers exist in request', function () {
      // r has no headers
      redirect.redirect(r);
      expect(r.testReturned).to.equal(404);
    });
    it('returns 404 if not all headers exist in request', function () {
      r.headersIn = {
        header3: 'value3',
        header1: 'value1',
      };
      redirect.redirect(r);
      expect(r.testReturned).to.equal(404);
    });
    it('returns 404 if all headers exist in request but not all match', function () {
      // r has all headers but not all values match
      r.headersIn = {
        header3: 'value3',
        header1: 'value1',
        header2: 'not-value2',
      };
      redirect.redirect(r);
      expect(r.testReturned).to.equal(404);
    });
    it('redirects to redirectPath if all headers exist in request and all match', function () {
      // r has all headers but not all values match
      r.headersIn = {
        header3: 'value3',
        header1: 'value1',
        header2: 'VALUE2', // header matching is case-insensitive. NGINX will lowercase the header name but not the value.
      };
      redirect.redirect(r);
      expect(r.testRedirectedTo).to.equal('/headers');
    });
  });
  describe('query parameter matching', function () {
    beforeEach(function () {
      let matches = {
        params: ['Arg1=value1', 'arg2=value2=SOME=other=value', 'arg3===value3&*1(*+'],
        redirectPath: '/params',
      };
      r.variables = {
        http_matches: JSON.stringify(matches),
      };
    });
    it('returns 404 if no args exist in request', function () {
      // r has no args
      redirect.redirect(r);
      expect(r.testReturned).to.equal(404);
    });
    it('returns 404 if not all args exist in request', function () {
      r.args = {
        Arg1: 'value1',
        arg2: 'value2=SOME=other=value',
      };
      redirect.redirect(r);
      expect(r.testReturned).to.equal(404);
    });
    it('returns 404 if all args exist in request but not all match', function () {
      // r has all headers but not all values match
      r.args = {
        Arg1: 'value1',
        arg2: 'value2=SOME=other=value',
        ARg3: '==value3&*1(*+', // query param matching is case sensitive, so this shouldn't match.
      };
      redirect.redirect(r);
      expect(r.testReturned).to.equal(404);
    });
    it('redirects to redirectPath if all args exist in request and all match', function () {
      // r has all headers but not all values match
      r.args = {
        Arg1: 'value1',
        arg2: 'value2=SOME=other=value',
        arg3: '==value3&*1(*+',
      };
      redirect.redirect(r);
      expect(r.testRedirectedTo).to.equal('/params');
    });
  });
  describe('multi-condition matching', function () {
    beforeEach(function () {
      let matches = {
        method: 'GET',
        headers: ['header1:value1', 'header2:value2'],
        params: ['Arg1=value1', 'arg2=value2=SOME=other=value'],
        redirectPath: '/a-match',
      };
      r.variables = {
        http_matches: JSON.stringify(matches),
      };
    });
    it('returns 405 if method does not match', function () {
      // method doesn't match but everything else does
      r.method = 'POST';
      r.headersIn = {
        header1: 'value1',
        header2: 'value2',
      };
      r.args = {
        Arg1: 'value1',
        arg2: 'value2=SOME=other=value',
      };
      redirect.redirect(r);
      expect(r.testReturned).to.equal(405);
    });
    it('returns 404 headers do not match', function () {
      // method and args match, but headers don't
      r.method = 'GET';
      r.headersIn = {
        header1: 'value1',
      };
      r.args = {
        Arg1: 'value1',
        arg2: 'value2=SOME=other=value',
      };
      redirect.redirect(r);
      expect(r.testReturned).to.equal(404);
    });
    it('returns 404 if args do not match', function () {
      // method and headers match, but args don't
      r.method = 'GET';
      r.headersIn = {
        header1: 'value1',
        header2: 'value2',
      };
      r.args = {
        arg2: 'value2=SOME=other=value',
      };
      redirect.redirect(r);
      expect(r.testReturned).to.equal(404);
    });
    it('redirects to redirectPath if all conditions match', function () {
      r.method = 'GET';
      r.headersIn = {
        header1: 'value1',
        header2: 'value2',
      };
      r.args = {
        Arg1: 'value1',
        arg2: 'value2=SOME=other=value',
      };
      redirect.redirect(r);
      expect(r.testRedirectedTo).to.equal('/a-match');
    });
  });
});
