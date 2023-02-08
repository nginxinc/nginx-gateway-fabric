/*
Package exchanger holds the CommandExchanger interface.

FIXME(kate-osborn): this package only holds one interface that is only used by the commander package.
It should be defined client-side, but the counterfeiter mock generator prevents this because of a cyclical import
cycle. Figure out a way to move this to the commander package.
*/
package exchanger
