// Copyright 2023 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package checker

import (
	"context"
	"net"
	"strings"
)

// NewVipCondition returns a new vip condition that checks whether the vip
// is bound to the given network interface named interfaceName.
//
// If interfaceName is empty, check all the network interfaces.
// If vip is empty, the condition always returns true.
func NewVipCondition(vip, interfaceName string) Condition {
	return ConditionFunc(func(ctx context.Context) (ok bool) {
		if vip == "" {
			return true
		}

		ok, err := ipIsOnInterface(vip, interfaceName)
		if err != nil {
			ok = false
		}
		return
	})
}

func ipIsOnInterface(ip, ifaceName string) (on bool, err error) {
	netip := net.ParseIP(strings.TrimSpace(ip))
	if netip == nil {
		return false, nil
	}

	var addrs []net.Addr
	var iface *net.Interface
	if ifaceName == "" {
		addrs, err = net.InterfaceAddrs()
	} else if iface, err = net.InterfaceByName(ifaceName); err == nil {
		addrs, err = iface.Addrs()
	}

	if err != nil {
		return
	}

	ip = netip.String()
	for _, addr := range addrs {
		if strings.Split(addr.String(), "/")[0] == ip {
			return true, nil
		}
	}

	return false, nil
}
