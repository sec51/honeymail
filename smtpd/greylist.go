/*

Store in a persistent storage a triplet:
- The IP address of the connecting host
- The envelope sender address
- The envelope recipient address(es), or just the first of them.

and also the timestamp of the first appearence

The email message will be dismissed with a temporary error until the configured period of time is elapsed
When a sender has proven itself able to properly retry delivery, it will be whitelisted for a longer period of time,
so that future delivery attempts will be unimpeded

*/

package smtpd
