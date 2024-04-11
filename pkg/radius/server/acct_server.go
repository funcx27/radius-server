package server

import (
	"fmt"
	"log"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
)

// 4	NAS-IP-Address	ipv4addr   #ocserv ip地址
// 6	Service-Type	enum        #不涉及
// 7	Framed-Protocol	enum         #不涉及
// 8	Framed-IP-Address	ipv4addr   #客户端分配的vpn内网地址
// 31	Calling-Station-Id	text       #客户端公网地址
// 40	Acct-Status-Type	enum   1 Start(计费开始) 2 Stop(计费结束) 3 Interium-Update(计费更新)
// 41	Acct-Delay-Time     integer   #计费数据报文延迟上报时间(s)
// 42	Acct-Input-Octets	integer   #客户端上传数据流量(bytes)  32位最大4GB count类型
// 43	Acct-Output-Octets	integer   #客户端下载数据流量(bytes)  32位最大4GB count类型
// 44	Acct-Session-Id	text          #客户端会话id
// 45	Acct-Authentic	enum         #不涉及
// 46	Acct-Session-Time	integer  #客户端连接时长(s)
// 47	Acct-Input-Packets	integer   #ocserv不支持
// 48   Acct-Output-Packets	integer   #ocserv不支持
// 49   Acct-Terminate-Cause enum      #客户端断开原因
// 52	Acct-Input-Gigawords	integer	[RFC2869] #Gbytes 32位 Acct-Input-Octets  溢出进位该字段
// 53	Acct-Output-Gigawords	integer [RFC2869] #Gbytes 32位 Acct-Output-Octets 溢出进位该字段

func (rs *RadiusServer) AcctServe(ListenAddr, metricsAddr string) {
	rs.Exporter = radiusExporter("0.0.0.0:9000")
	server := radius.PacketServer{
		Handler:      radius.HandlerFunc(rs.acctHandler()),
		SecretSource: radius.StaticSecretSource([]byte(rs.secret)),
	}
	if err := serve(&server, ListenAddr); err != nil {
		log.Fatal(err)
	}
}

func (rs *RadiusServer) acctHandler() func(w radius.ResponseWriter, r *radius.Request) {
	return func(w radius.ResponseWriter, r *radius.Request) {
		as := AccountingSession{
			SessionId:       rfc2866.AcctSessionID_GetString(r.Packet),
			Username:        rfc2865.UserName_GetString(r.Packet),
			NasIdentifier:   rfc2865.NASIdentifier_GetString(r.Packet),
			Status:          rfc2866.AcctStatusType_Get(r.Packet),
			SessionTime:     uint32(rfc2866.AcctSessionTime_Get(r.Packet)),
			RemoteIpAddress: rfc2865.CallingStationID_GetString(r.Packet),
			FramedIPAddress: rfc2865.FramedIPAddress_Get(r.Packet),
			InputBytes:      uint64(rfc2866.AcctInputOctets_Get(r.Packet)) + uint64(rfc2869.AcctInputGigawords_Get(r.Packet))*1<<30,
			OutputBytes:     uint64(rfc2866.AcctOutputOctets_Get(r.Packet)) + uint64(rfc2869.AcctOutputGigawords_Get(r.Packet))*1<<30,
			TerminateCause:  rfc2866.AcctTerminateCause_Get(r.Packet),
		}
		if as.Status > rfc2866.AcctStatusType_Value_Start {
			fmt.Println(as)
			labelValues := []string{as.SessionId, as.RemoteIpAddress, as.Username, as.NasIdentifier, as.FramedIPAddress.String()}
			rs.WriteMetrics("ocserv", "user_session", "status", float64(as.Status), labelValues...)
			rs.WriteMetrics("ocserv", "user_session", "uptimes", float64(as.SessionTime), labelValues...)
			rs.WriteMetrics("ocserv", "user_session", "output_bytes", float64(as.OutputBytes), labelValues...)
			rs.WriteMetrics("ocserv", "user_session", "input_bytes", float64(as.InputBytes), labelValues...)
			rs.WriteMetrics("ocserv", "user_session", "terminate_cause", float64(as.TerminateCause), labelValues...)
		}

		promLabels := prometheus.Labels{
			"id":             as.SessionId,
			"remote_address": as.RemoteIpAddress,
			"framed_address": as.FramedIPAddress.String(),
			"username":       as.Username,
			"nasidentifier":  as.NasIdentifier,
		}
		if as.Status == rfc2866.AcctStatusType_Value_Stop || as.Status > rfc2866.AcctStatusType_Value_InterimUpdate {
			for namespace, subsystems := range rs.Exporter.MetricsMetaData {
				for subsystem, names := range subsystems {
					for name := range names {
						if !rs.DeleteMetrics(string(namespace), string(subsystem), string(name), promLabels) {
							fmt.Println(strings.Join([]string{string(namespace), string(subsystem), string(name)}, "_"), "delete failed:", promLabels)
						}
					}
				}
			}
		}
		w.Write(r.Response(radius.CodeAccountingResponse))
	}
}
