#!/usr/bin/env bash
#
# Copyright (C) 2014-2019 LEAP
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#
# bash translation of bitmask-root
#
# USAGE:
#   bitmask-root firewall stop
#   bitmask-root firewall start [restart] GATEWAY1 GATEWAY2 ...
#   bitmask-root openvpn stop
#   bitmask-root openvpn start CONFIG1 CONFIG1 ...
#
# All actions return exit code 0 for success, non-zero otherwise.
#
# The `openvpn start` action is special: it calls exec on openvpn and replaces
# the current process. If the `restart` parameter is passed, the firewall will
# not be teared down in the case of an error during launch.

set -e

if [ -n "${DEBUG}${TEST}" ]; then
	set -x
fi

VERSION=19
SCRIPT=${0##*/}

IFS=$' \t\n'
REAL_IFS=$IFS # copy to restore later when changing IFS
if [ "$UDP" = "1" ]; then
	NAMESERVER=10.42.0.1
else
	NAMESERVER=10.41.0.1
fi
CHAIN=bitmask
CHAIN_NAT=bitmask
CHAIN_POST=bitmask_postrouting

LEAPOPENVPN=LEAPOPENVPN
QUBES_PROXY=false
USER_ID=$(id -u)
OPENVPN_GROUP=""
for g in nobody nogroup; do
	if getent group "$g" >/dev/null 2>&1; then
		OPENVPN_GROUP=$g
		break
	fi
done

# Are these needed at all?
OPENVPN_USER=nobody
LOCAL_INTERFACE=lo

log_msg() {
	>&2 echo "$SCRIPT: $1"
	logger -p "daemon.${2:-info}" -t "$SCRIPT" -- "$1" 2>/dev/null || :
}

bail() {
	log_msg "$1" err
	exit 1
}

# the python script checks for a literal file in "/bin", "/sbin", "/usr/bin", "/usr/sbin", "/usr/local/sbin"
# the order is flipped, normally you want to check in "/usr/local" "/usr" "/" instead
# also checking for a literal file is not correct, you need to make sure it has
# executable permission, the command -v builtin of shell does all of this for us
swhich() {
	p=/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin
	PATH=$p command -v "$1" || bail "Can't find $1"
}

IP=$(swhich ip)
IPTABLES=$(swhich iptables)
IP6TABLES=$(swhich ip6tables)
SYSCTL=$(swhich sysctl)

QUBES_VER=0
if [ -f /var/run/qubes/this-is-proxyvm ] && [ -d /etc/qubes ]; then
	QUBES_CFG=/rw/config/
	QUBES_IPHOOK=${QUBES_CFG}qubes-ip-change-hook
	QUBES_FW_SCRIPT=${QUBES_CFG}qubes-firewall-user-script
	if "$IPTABLES" --list QBS-FORWARD >/dev/null 2>&1; then
		QUBES_VER=4
	else
		QUBES_VER=3
	fi
fi

usage() {
	cat <<-EOF
	This is $SCRIPT version $VERSION

	This program manipulates the Bitmask firewall. It is *NOT* intented to be used manually.

	Commands:

	$SCRIPT version
	$SCRIPT restart
	$SCRIPT openvpn start <args>
	$SCRIPT openvpn stop
	$SCRIPT firewall start <args>
	$SCRIPT firewall stop
	$SCRIPT firewall isup
	EOF
	exit 0
}

check_root() {
	[ "$USER_ID" = 0 ] || bail "ERROR: must be run as root"
}

# The python script uses sysctl -a 2>/dev/null | grep all.disable_ipv6 | grep 1
# which is flawed. the sysctl setting can only exists when the IPv6 stack was initialized
# so if you are running a kernel with ipv6 just not compiled in or with the ipv6.disable=1
# kernel parameter, the old method will return a false positive!
#
# Instead, lets check for the presence of the sysctl key and read it
# This is also faster than shelling out sysctl | grep | grep
is_ipv6_disabled() {
	f=/proc/sys/net/ipv6/conf/all/disable_ipv6
	if [ ! -f "$f" ];        then return 0       # disabled at kernel level
	elif read -r val < "$f"; then [ "$val" = 1 ] # disabled by sysctl
	else                     return 1            # should never be reached
	fi
}

# python uses socket.inet_aton, here we have to shell out grep -E technically a
# bunch of case statements with regexes would remove grep but that becomes unreadable mess
is_valid_address() {
	printf '%s' "$1" | grep -qEx '[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}' || return 1
	# make sure no number is greater than 255
	IFS=.
	set -- $1
	IFS=$REAL_IFS
	for octet do
		if [ "$octet" -gt 255 ]; then
			return 1
		fi
	done
	return 0
}

# python uses ipaddress to validate which is much simpler
getIPv4AllowAddresses() {
	#    10.0.0.0/8           172.16.0.0/12           192.168.0.0/16    27.0.0.0/8
	# `|| :` so a missing file yields empty instead of aborting on
	# awk's non-zero exit when the file is absent (python tolerates it)
	awk '/^10\./ || /^172\.(1[6-9]|2[0-9]|3[01])\./ || /^192\.168\./ || /^127\./' /etc/bitmask/ipv4.allow 2>/dev/null || :
}

getIPv6AllowAddresses() {
	#    fc00::/7                   fe80::/10                      ::1
	awk '/^[[:space:]]*(fc|fd)/ || /^[[:space:]]*::1/ || /^[[:space:]]*[fF][eE][89aAbB]/' /etc/bitmask/ipv6.allow 2>/dev/null || :
}

run_iptable_with_check_flag() {
	mode=""
	for a do
		case "$a" in
			--append|--insert) mode=add;;
			--delete)          mode=del;;
		esac
	done

	if [ "$mode" = "" ]; then
		"$@"
		return $?
	fi

	set +e
	# check the entire array and swap --{append,insert,delete} for --check
	# before executing with it, needs to be done in a subshell
	# since we need to preserve the original array for later
	(
		for a do
			case "$a" in
				--append|--insert|--delete) a=--check;;
			esac
			set -- "$@" "$a"
			shift
		done
		"$@" >/dev/null 2>&1
	)
	status=$?
	set -e

	if   [ "$mode" = add ] && [ "$status" != 0 ]; then "$@"
	elif [ "$mode" = del ] && [ "$status" = 0 ];  then "$@"
	fi
}

iptables_both() {
	# Run on both ipv4 and ipv6 iptables.
	ip4tables "$@"
	ip6tables "$@"
}

ip4tables() {
	run_iptable_with_check_flag "$IPTABLES" "$@"
}

ip6tables() {
	run_iptable_with_check_flag "$IP6TABLES" "$@"
}

ipv4_chain_exists() (
	chain="$1" table="$2"
	set +e
	if [ -n "$table" ]; then
		"$IPTABLES" -t "$table" --list "$chain" --numeric >/dev/null 2>&1
	else
		"$IPTABLES" --list "$chain" --numeric >/dev/null 2>&1
	fi
	status=$?
	set -e
	case $status in
		0) return 0;;
		1) return 1;;
		*) log_msg "ERROR: Could not determine state of iptable chain"
			return 1
			;;
	esac
)

ipv6_chain_exists() (
	chain="$1"
	set +e
	"$IP6TABLES" --list "$chain" --numeric >/dev/null 2>&1
	status=$?
	set -e
	case $status in
		0) return 0;;
		1) return 1;;
		*) log_msg "ERROR: Could not determine state of ip6table chain"
			return 1
			;;
	esac
)

get_openvpn_bin() {
	# /snap/bin/riseup-vpn.openvpn no longer searched since the snap was dropped right?
	set -- \
		/usr/sbin/openvpn \
		/usr/local/sbin/leap-openvpn

	for bin do
		if command -v "$bin"; then
			return 0
		fi
	done

	bail "Could not find openvpn binary"
}

_validate_openvpn_param() {
	passed=0
	type_spec=$1
	value=$2
	IFS='|'
	for type_part in $type_spec; do
		case "$type_part" in
			'')
				continue
				;;
			NUMBER)
				case "$value" in
					''|*[!0-9]*) passed=0;;
					*)           passed=1;;
				esac
				;;
			PROTO)
				case "$value" in
					tcp|udp|tcp4|udp4) passed=1;;
				esac
				;;
			IP)
				if is_valid_address "$value"; then
					passed=1
				fi
				;;
			CIPHER)
				case "$value" in
					*[!A-Za-z0-9:\-]*) passed=0;;
					*)                 passed=1;;
				esac
				;;
			USER)
				case "$value" in
					[a-zA-Z0-9_.@][a-zA-Z0-9_\-\.@]*\$|[a-zA-Z0-9_.@][a-zA-Z0-9_\-\.@]*)
						passed=1
						;;
				esac
				;;
			FILE)
				if [ -f "$value" ]; then
					passed=1
				fi
				;;
			DIR)
				if [ -d "${value%/*}" ]; then
					passed=1
				fi
				;;
			UNIXSOCKET)
				if [ "$value" = unix ]; then
					passed=1
				fi
				;;
			NETGW)
				if [ "$value" = net_gateway ]; then
					passed=1
				fi
				;;
			UID)
				case "$value" in
					*[!a-zA-Z0-9]*) passed=0;;
					*)              passed=1;;
				esac
				;;
			LOGFILE)
				if [ "$value" = /tmp/leap-vpn.log ]; then
					passed=1
				fi
				;;
			ignore)
				if [ "$value" = ignore ]; then
					passed=1
				fi
				;;
			route)
				if [ "$value" = route ]; then
					passed=1
				fi
				;;
		esac
		# if one alternative matched, no need to check the rest
		if [ "$passed" = 1 ]; then
			break
		fi
	done

	IFS=$REAL_IFS
	if [ "$passed" = 1 ]; then
		return 0
	fi
	return 1
}

# echo the required params for a given flag name, or "__NONE__" if no params.
_get_allowed_flag_info() {
	case "$1" in
		--remote)                 echo "IP NUMBER PROTO";;
		--tls-cipher)             echo "CIPHER";;
		--cipher)                 echo "CIPHER";;
		--auth)                   echo "CIPHER";;
		--management)             echo "DIR||IP UNIXSOCKET||NUMBER FILE";;
		--management-client-user) echo "USER";;
		--route)                  echo "IP IP NETGW";;
		--cert)                   echo "FILE";;
		--key)                    echo "FILE";;
		--ca)                     echo "FILE";;
		--fragment)               echo "NUMBER";;
		--keepalive)              echo "NUMBER NUMBER";;
		--verb)                   echo "NUMBER";;
		--management-client)      echo "__NONE__";;
		--tun-ipv6)               echo "__NONE__";;
		--log)                    echo "LOGFILE";;
		--pull-filter)            echo "ignore route";;
		--socks-proxy)            echo "IP NUMBER";;
		*)                        return 1;;
	esac
}

parse_openvpn_flags() {
	_flags=()
	while :; do
		[ $# -eq 0 ] && break
		flag=$1
		shift

		case "$flag" in
			--*) :;;
			*) continue;;
		esac

		if ! allowed=$(_get_allowed_flag_info "$flag"); then
			log_msg "WARNING: unrecognized openvpn flag $flag"
			continue
		fi

		_params=()
		while [ $# -gt 0 ]; do
			case "$1" in
				--*) break;;
			esac
			_params+=("$1")
			shift
		done

		if [ "$allowed" = __NONE__ ]; then
			if [ ${#_params[@]} != 0 ]; then
				log_msg "ERROR: $flag takes no params"
				return 1
			fi
		else
			_specs=($allowed)
			if [ ${#_specs[@]} != ${#_params[@]} ]; then
				log_msg "ERROR: wrong param count for $flag"
				return 1
			fi

			for i in "${!_specs[@]}"; do
				if ! _validate_openvpn_param "${_specs[i]}" "${_params[i]}"; then
					log_msg "ERROR: Bad argument ${_params[i]}"
					return 1
				fi
			done
		fi

		_flags+=("$flag" "${_params[@]}")
	done
	# emit one arg per line so the caller can re-assemble them as an
	# array. joining them with spaces and word-splitting later would let
	# spaces/globs in a file or address leak into multiple argv entries.
	printf '%s\n' "${_flags[@]}"
}

openvpn_start() {
	if ! safe_flags=$(parse_openvpn_flags "$@") || [ -z "$safe_flags" ]; then
		bail "ERROR: could not parse openvpn options"
	fi
	readarray -t safe_flags <<<"$safe_flags"

	OPENVPN=$(get_openvpn_bin)

	set -- \
	  --setenv LEAPOPENVPN 1 --nobind --client --dev tun        \
	  --tls-client --remote-cert-tls server --management-signal \
	  --script-security 1 --user nobody --auth-nocache --tls-version-min 1.2

	if [ -n "$OPENVPN_GROUP" ]; then
		set -- "$@" --group "$OPENVPN_GROUP"
	fi

	if is_ipv6_disabled; then
		set -- "$@" \
		  --pull-filter ignore ifconfig-ipv6 --pull-filter ignore route-ipv6
	fi

	exec "$OPENVPN" "$@" "${safe_flags[@]}"
}

# the python script manually iterates over /proc/*/cmdline
# we could do the same with a for loop but
# then we would have to mess with null bytes in shell and that's a mess
#
# this also does not need to be a function, made it for reference to the python script
get_process_list() {
	ps -eo pid=,args=
}

openvpn_stop() {
	# Stop the openvpn that has likely been launched by bitmask.
	get_process_list | while read -r pid cmdline; do
		case "$cmdline" in
			*openvpn*"$LEAPOPENVPN"*|*"$LEAPOPENVPN"*openvpn*)
				kill -TERM "$pid" 2>/dev/null || :
				break
				;;
		esac
	done
}

###############################################################################
# FIREWALL HELPERS
###############################################################################

get_gateways() {
	# Filter a passed list of gateways, returning only the valid ones.
	# Outputs valid gateways one per line.
	found=0
	for gw in "$@"; do
		if is_valid_address "$gw"; then
			echo "$gw"
			found=1
		fi
	done
	[ "$found" != 0 ] || return 1
	return 0
}

get_default_device() {
	# Retrieve the current default network device.
	routes="$("$IP" route show 2>/dev/null)" || return 1
	device="${routes#*default*dev }"
	device="${device%% *}"
	[ -n "$device" ] || return 1
	echo "$device"
}

get_local_network_ipv4() {
	device=$1
	addresses=$("$IP" -o address show dev "$device" 2>/dev/null) || return 1
	case "$addresses" in
		*'inet '*)
			addr=${addresses#*inet }
			addr=${addr%% *}
			;;
		*)
			return 1
			;;
	esac
	echo "$addr"
}

get_local_network_ipv6() {
	device=$1
	addresses=$("$IP" -o address show dev "$device" 2>/dev/null) || return 1
	case "$addresses" in
		*'inet6 '*)
			addr=${addresses#*inet6 }
			addr=${addr%% *}
			;;
		*)
			return 1
			;;
	esac
	echo "$addr"
}

enable_ip_forwarding() {
	echo "1" > /proc/sys/net/ipv4/ip_forward 2>/dev/null || :
}

_firewall_start_commands() {
	default_device="$(get_default_device)" || bail "Could not find default device"
	local_network_ipv4=$(get_local_network_ipv4 "$default_device") || :
	local_network_ipv6=$(get_local_network_ipv6 "$default_device") || :
	gateways=$(get_gateways "$@") || bail "ERROR: No valid gateways specified"

	# Create custom chains
	ipv4_chain_exists "$CHAIN" || ip4tables --new-chain "$CHAIN"
	ipv4_chain_exists "$CHAIN_NAT" nat || ip4tables --table nat --new-chain "$CHAIN_NAT"
	ipv4_chain_exists "$CHAIN_POST" nat || ip4tables --table nat --new-chain "$CHAIN_POST"
	ipv6_chain_exists "$CHAIN" || ip6tables --new-chain "$CHAIN"

	ip4tables --table nat --insert OUTPUT --jump "$CHAIN_NAT"
	ip4tables --table nat --insert POSTROUTING --jump "$CHAIN_POST"
	iptables_both --insert OUTPUT --jump "$CHAIN"
}

_firewall_start_ipv4() {
	ipv4_exceptions=$(getIPv4AllowAddresses)
	[ -n "$local_network_ipv4" ] || return 0

	if [ -z "$ipv4_exceptions" ]; then
		# Allow all network destinations if no explicit allow list
		ip4tables --append "$CHAIN" \
			--destination "$local_network_ipv4" -o "$default_device" --jump ACCEPT
	fi
	# Allow network sources for DNS
	ip4tables --append "$CHAIN" \
		--source "$local_network_ipv4" -o "$default_device" -p udp --dport 53 --jump ACCEPT
	ip4tables --append "$CHAIN" \
		--source "$local_network_ipv4" -o "$default_device" -p tcp --dport 53 --jump ACCEPT
	# Allow multicast SSDP
	ip4tables --append "$CHAIN" \
		--protocol udp --destination 239.255.255.250 --dport 1900 \
		-o "$default_device" --jump RETURN
	# Allow multicast mDNS
	ip4tables --append "$CHAIN" \
		--protocol udp --destination 224.0.0.251 --dport 5353 \
		-o "$default_device" --jump RETURN
}

_firewall_start_ipv6() {
	ipv6_exceptions="$(getIPv6AllowAddresses)"
	[ -n "$local_network_ipv6" ] || return 0
	if [ -z "$ipv6_exceptions" ]; then
		ip6tables --append "$CHAIN" \
			--destination "$local_network_ipv6" -o "$default_device" --jump ACCEPT
	fi
	ip6tables --append "$CHAIN" --protocol udp --destination "FF05::C" \
		--dport 1900 -o "$default_device" --jump RETURN
	ip6tables --append "$CHAIN" --protocol udp --destination "FF02::FB" \
		--dport 5353 -o "$default_device" --jump RETURN
}

_firewall_start_qubes() {
	[ "$QUBES_VER" -ge 3 ] || return 0
	if ! grep -q "installed by $SCRIPT" "$QUBES_FW_SCRIPT" 2>/dev/null; then
		cat > "$QUBES_FW_SCRIPT" <<-EOF
		#!/bin/sh
		# Anti-leak rules installed by $SCRIPT $VERSION
		iptables  --insert FORWARD -i eth0 -j DROP
		iptables  --insert FORWARD -o eth0 -j DROP
		ip6tables --insert FORWARD -i eth0 -j DROP
		ip6tables --insert FORWARD -o eth0 -j DROP
		iptables  --insert INPUT -i tun+ -j DROP
		ip6tables --insert INPUT -i tun+ -j DROP
		EOF
		chmod 0700 "$QUBES_FW_SCRIPT"
		if [ ! -e "$QUBES_IPHOOK" ]; then
			ln -s "$QUBES_FW_SCRIPT" "$QUBES_IPHOOK" 2>/dev/null || :
		fi
		if [ "$QUBES_VER" = 4 ]; then
			"$QUBES_FW_SCRIPT"
		elif [ "$QUBES_VER" = 3 ]; then
			systemctl restart qubes-firewall.service >/dev/null 2>&1 || :
		fi
	fi
}

firewall_start() {
	# Bring up the firewall.
	# Args: [restart] gateway1 gateway2 ...
	#
	# If "restart" is passed, the firewall is not torn down on error
	# so it can be retried without losing existing rules.

	_RESTART=0
	for arg do
		case "$arg" in
			restart) _RESTART=1; shift;;
			*) set -- "$@" "$arg"; shift;;
		esac
	done
	trap '[ "$_RESTART" -eq 0 ] && firewall_stop || :' EXIT

	_firewall_start_commands "$@"

	# Route all ipv4 DNS over VPN
	enable_ip_forwarding

	if [ "$QUBES_VER" -ge 3 ]; then
		# Qubes rewrites DNS
		ip4tables -t nat --flush PR-QBS
		echo "$gateways" | while IFS= read -r gw; do
			[ -n "$gw" ] || continue
			ip4tables -t nat --append PR-QBS --destination "$gw" --jump RETURN
		done
		ip4tables -t nat --append PR-QBS -p udp --dport 53 --jump DNAT --to "${NAMESERVER}:53"
		ip4tables -t nat --append PR-QBS -p tcp --dport 53 --jump DNAT --to "${NAMESERVER}:53"
	else
		# Normal DNS rewrite
		echo "$gateways" | while IFS= read -r gw; do
			[ -n "$gw" ] || continue
			ip4tables -t nat --append "$CHAIN_NAT" --destination "$gw" --jump RETURN
		done

		ip4tables -t nat --append "$CHAIN_NAT" --protocol udp \
			--dest "127.0.1.1,127.0.0.1,127.0.0.53" --dport 53 --jump ACCEPT

		ip4tables -t nat --append "$CHAIN_NAT" -p udp --dport 53 \
			--jump DNAT --to "${NAMESERVER}:53"
		ip4tables -t nat --append "$CHAIN_NAT" -p tcp --dport 53 \
			--jump DNAT --to "${NAMESERVER}:53"

		# Masquerade for rewritten DNS packets
		ip4tables -t nat --append "$CHAIN_POST" \
			--dest "$NAMESERVER" --protocol udp --dport 53 --jump MASQUERADE
		ip4tables -t nat --append "$CHAIN_POST" \
			--dest "$NAMESERVER" --protocol tcp --dport 53 --jump MASQUERADE
	fi

	_firewall_start_ipv4
	_firewall_start_ipv6

	# ---- Allow traffic to gateways ----
	echo "$gateways" | while IFS= read -r gw; do
		[ -n "$gw" ] || continue
		ip4tables --append "$CHAIN" --destination "$gw" -o "$default_device" --jump ACCEPT
	done

	# ---- Debug logging ----
	if [ -n "$DEBUG" ]; then
		iptables_both --append "$CHAIN" -o "$default_device" \
			--jump LOG --log-prefix "iptables denied: " --log-level 7
	fi

	# ---- Explicit private exceptions ----
	if [ -n "$ipv4_exceptions" ]; then
		echo "$ipv4_exceptions" | while IFS= read -r ip; do
			[ -n "$ip" ] || continue
			ip4tables --append "$CHAIN" --destination "$ip" -o "$default_device" --jump ACCEPT
		done
		ip4tables --append "$CHAIN" \
			--destination "$local_network_ipv4" -o "$default_device" --jump REJECT
	fi

	if [ -n "$ipv6_exceptions" ]; then
		echo "$ipv6_exceptions" | while IFS= read -r ip; do
			[ -n "$ip" ] || continue
			ip6tables --append "$CHAIN" --destination "$ip" -o "$default_device" --jump ACCEPT
		done
		ip6tables --append "$CHAIN" \
			--destination "$local_network_ipv6" -o "$default_device" --jump REJECT
	fi

	# ---- Reject all other IPv6 ----
	ip6tables --append "$CHAIN" -p tcp --jump REJECT
	ip6tables --append "$CHAIN" -p udp --jump REJECT

	# ---- Reject all other IPv4 over default device ----
	ip4tables --append "$CHAIN" -o "$default_device" --jump REJECT

	# ---- Qubes anti-leak rules ----
	_firewall_start_qubes

	trap - EXIT
}

_firewall_stop_commands() {
	iptables_both --delete OUTPUT --jump "$CHAIN" || status=$?
	ip4tables -t nat --delete OUTPUT --jump "$CHAIN_NAT" || status=$?
	ip4tables -t nat --delete POSTROUTING --jump "$CHAIN_POST" || status=$?
	ip4tables --flush "$CHAIN" && ip4tables --delete-chain "$CHAIN"  || status=$?
	ip4tables -t nat --flush "$CHAIN_NAT" && ip4tables -t nat --delete-chain "$CHAIN_NAT" || status=$?
	ip4tables -t nat --flush "$CHAIN_POST" && ip4tables -t nat --delete-chain "$CHAIN_POST" || status=$?
	ip6tables --flush "$CHAIN" && ip6tables --delete-chain "$CHAIN" || status=$?
}

firewall_stop() {
	status=0
	_firewall_stop_commands

	# NOTE: the python script has a bug here:
	# it checks `if not (ok or ipv4_chain_exists or ipv6_chain_exists)`
	# without`()` on the function calls,
	# so the exception is never raised and failures are silently ignored.
	if [ "$status" = 0 ]; then
		return 0
	elif ipv4_chain_exists "$CHAIN" || ipv6_chain_exists "$CHAIN"; then
		log_msg "firewall might still be left up. Please try 'firewall stop' again." err
	fi
	return 1
}


case "$1" in
	--help|help|-h)
		usage
		;;
	--version|version)
		echo "$VERSION"
		exit 0
		;;
	openvpn)
		shift
		check_root
		case "$1" in
			start) shift; openvpn_start "$@";;
			stop)  shift; openvpn_stop;;
			*)     bail "ERROR: No such command. Try $SCRIPT help";;
		esac
		;;
	firewall)
		shift
		check_root
		case "$1" in
			start) shift; firewall_start "$@";;
			stop)  shift; firewall_stop || bail "ERROR: could not stop firewall";;
			isup)  shift; ipv4_chain_exists "$CHAIN" || bail "INFO: bitmask firewall is down";;
			*)     bail "ERROR: No such command. Try $SCRIPT help";;
		esac;;
	*)
		bail "ERROR: No such command. Try $SCRIPT help"
		;;
esac
