Some design notes about gateway selection
=========================================

These are some notes about the design for the new gateway selector.

1. Default behavior: automatic gateway selection.
-------------------------------------------------

This is what every user experiences by default. No configuration, we just try
to find the "best" gateway for a given location.

Client tries to fetch a recommended gateway list from menshen, and pass
a subset of `n` gateways to openvpn. Openvpn will then retry the remotes in
that order if any of those fails. For the desktop client, `n=3`.

1.1 If new-menshen is deployed (at riseup, for the time being), we get
    `sortedGateways` field, with load metrics. Menshen has already taken our
    geolocated ip into account to  first sort by distance and then by load.

1.2 If old-menshen (ie, geoip) is available, we fetch the `gateways` field

    curl -k https://api.black.riseup.net:9001/json

1.3 If no menshen is available (service down, unreachable etc) or the results
    are unusable (misconfigured provider, geolocated gateway fails to assign
    lat/lon coordinates), we fallback to the timezone distance heuristic.

    NOTE: we should catch the failure to geolocate.
    NOTE: we could inform the user of the timezone heuristic, if done, for transparency.


2. Give connection feedback to the user
----------------------------------------

Little is done right now about this, but we want to give feedback to the user.

* We can display the IP that menshen sees before connecting (your ip: x.x.x.x, your country: xx).
* When connecting, we could display the openvpn states (so that user see changes).
* When vpn connects to a gateway, we get the ip for the gateway (from the
  openvn management interface), and match that IP against our list of gateways.
  We use this to detect which gw we're currently connected to, and display the
  location for the gateway we're connected to.

* info to display: (where am I exiting through?)

  - city
  - country
  - ip


* systray: RiseupVPN on / Paris (auto)


3. manual gateway override
--------------------------

The UI should offer the user a way to manually override gateways, at the level of cities.
Ideally there's a toggle near the indication of automatic/manual for gateways (systray or window).

4. gateway selector
-------------------

If user wants to do manual override, we display another panel, with

- a list of available locations
- some health indicator for those nodes, if available.

Open questions: 

- how hard is to add an icon to a Qt combobox?
- how much info is enough/too much? (like: does user need to know gateway name? probably not. but useful for logs at least).
- do we want to expose the gateway/transport level to the user? (probably not).


5. picking host for a manual override
-------------------------------------

when user select a new location as part of manual override, we select the best gateway for that location:
    
1 **if menshen is available**, we just pick the first in the ordered subset of gws for that location

- *question: should we still pass all the other more congested gateways for that location as fallback remotes to openvpn? or just the one?*

2 **if menshen is not available**, we just pick randomly from the gws in a given city.

- *same question as in point above*.
    
6. user feedback for new connection
-----------------------------------

After a manual gateway override, we try to connect (ideally do some connectivity/dns tests), and then change the exit location.

We now display the location for the new gateway:

* systray: RiseupVPN on / Paris (manual)


7. revert to automatic choice
-----------------------------

Selecting automatic gw selection again should be a simple action, available quickly from the systray or the main window.

* systray (clickable item):

✓ Automatic

(maybe this should be called *fastest* instead?


More advanced use cases
=======================

A1 **how to display load accurately**

(load indication can be averaged, or best-case etc...) for every city

A2 **refresh info in the backgroud**

There are some advantages in refreshing the recommended gateways in the
background:

- no need to spend time fetching that info on next reconnect
- ability to signal user if we're currently in a very congested gateway
- ability to transparently switch to another gateway for the same location on next
reconnect

...but also some disadvantages:

- currently, there's an autoincrement in lb that we have to solve before doing this refresh.
- once we're connected to the gateway, menshen will not see our "original" ip,
  so the geolocation info will not be valid anymore (unless we pass city or
  coordinates as another parameter).

What to do when/if we detect we're connected to a congested / bad gateway?

- consensus now: take the next reconnect opportunity to change gws.
- maybe ask for user confirmation if that change means breaking expectations (gw in a different country/state for instance).

A3 **inform menshen of manual override**.

We agreed that it would be great to have an estimation of how many users are doing manual overrides.

For this, a *proposal* is to add an endpoint to menshen in which the periodic
refresh passes the manual override as a parameter. Something like:

GET menshen.float.ip/json?city=paris&type=refresh

A4 **how to gracefully fail**

(ie, if menshen geoip cannot give lat/lon for your current ip, or if the geoip service is down/blocked) -> we have the "timezone heuristic", it would be good to be explicit about this choice.

- detect dns failures

A4. advanced visualization: map / traffic graphs - yayyy eye candy :) we want that stuff, but needs more research.

