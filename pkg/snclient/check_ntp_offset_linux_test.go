package snclient

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckNTPOffset(t *testing.T) {
	snc := StartTestAgent(t, "")

	res := snc.RunCheck("check_ntp_offset", []string{"warn=offset >= 10000", "crit=offset >= 20000"})
	if res.State == CheckExitUnknown || res.State == CheckExitCritical {
		t.Logf("skipped, no working ntp detected")
		StopTestAgent(t, snc)
		t.Skip()

		return
	}

	assert.Equalf(t, CheckExitOK, res.State, "state OK")
	assert.Containsf(t, string(res.BuildPluginOutput()), "OK - offset", "output matches")

	StopTestAgent(t, snc)
}

func TestCheckNTPOffsetTimeDateCtl(t *testing.T) {
	snc := StartTestAgent(t, "")

	// mock timedatectl command from output of: timedatectl timesync-status
	tmpPath := MockSystemUtilities(t, map[string]string{
		"timedatectl": `Server: 62.225.132.250 (0.debian.pool.ntp.org)
Poll interval: 17min 4s (min: 32s; max 34min 8s)
         Leap: normal
      Version: 4
      Stratum: 2
    Reference: C035676C
    Precision: 1us (-22)
Root distance: 47.041ms (max: 5s)
       Offset: -32.316ms
        Delay: 30.801ms
       Jitter: 236.187ms
 Packet count: 14
    Frequency: +49.094ppm`,
	})
	defer os.RemoveAll(tmpPath)
	res := snc.RunCheck("check_ntp_offset", []string{"source=timedatectl"})
	assert.Equalf(t, CheckExitOK, res.State, "state OK")
	assert.Equalf(t, "OK - offset -32.3ms from 62.225.132.250 (0.debian.pool.ntp.org) |'offset'=-32.316ms;-50:50;-100:100 'stratum'=2;;;0 'jitter'=236.187ms;;;0",
		string(res.BuildPluginOutput()), "output matches")

	res = snc.RunCheck("check_ntp_offset", []string{}) // with source=auto
	assert.Equalf(t, CheckExitOK, res.State, "state OK")
	assert.Equalf(t, "OK - offset -32.3ms from 62.225.132.250 (0.debian.pool.ntp.org) |'offset'=-32.316ms;-50:50;-100:100 'stratum'=2;;;0 'jitter'=236.187ms;;;0",
		string(res.BuildPluginOutput()), "output matches")

	// mock critical response
	MockSystemUtilities(t, map[string]string{
		"timedatectl": `Server: 62.225.132.250 (0.debian.pool.ntp.org)
Poll interval: 17min 4s (min: 32s; max 34min 8s)
         Leap: normal
      Version: 4
      Stratum: 2
    Reference: C035676C
    Precision: 1us (-22)
Root distance: 47.041ms (max: 5s)
       Offset: -132.316ms
        Delay: 30.801ms
       Jitter: 236.187us
 Packet count: 14
    Frequency: +49.094ppm`,
	})
	res = snc.RunCheck("check_ntp_offset", []string{"source=timedatectl"})
	assert.Equalf(t, CheckExitCritical, res.State, "state Critical")
	assert.Equalf(t, "CRITICAL - offset -132ms from 62.225.132.250 (0.debian.pool.ntp.org) |'offset'=-132.316ms;-50:50;-100:100 'stratum'=2;;;0 'jitter'=0.236187ms;;;0",
		string(res.BuildPluginOutput()), "output matches")

	StopTestAgent(t, snc)
}

func TestCheckNTPOffsetChronyc(t *testing.T) {
	snc := StartTestAgent(t, "")

	// mock chronyc command from output of: chronyc tracking
	tmpPath := MockSystemUtilities(t, map[string]string{
		"chronyc": `Reference ID    : B9FC8C7D (test.ntp)
Stratum         : 3
Ref time (UTC)  : Thu Dec 21 21:46:26 2023
System time     : 0.003751489 seconds fast of NTP time
Last offset     : +0.002065938 seconds
RMS offset      : 0.002065938 seconds
Frequency       : 25.960 ppm fast
Residual freq   : +0.000 ppm
Skew            : 109.671 ppm
Root delay      : 0.038340077 seconds
Root dispersion : 0.002839299 seconds
Update interval : 64.4 seconds
Leap status     : Normal`,
	})
	defer os.RemoveAll(tmpPath)
	res := snc.RunCheck("check_ntp_offset", []string{"source=chronyc"})
	assert.Equalf(t, CheckExitOK, res.State, "state OK")
	assert.Equalf(t, "OK - offset 2.066ms from test.ntp |'offset'=2.065938ms;-50:50;-100:100 'stratum'=3;;;0",
		string(res.BuildPluginOutput()), "output matches")

	MockSystemUtilities(t, map[string]string{
		"chronyc": `Reference ID    : B9FC8C7D (test.ntp)
Stratum         : 3
Ref time (UTC)  : Thu Dec 21 22:01:32 2023
System time     : 0.000939555 seconds fast of NTP time
Last offset     : -0.019212097 seconds
RMS offset      : 0.001114742 seconds
Frequency       : 0.992 ppm slow
Residual freq   : -0.024 ppm
Skew            : 1.524 ppm
Root delay      : 0.037130691 seconds
Root dispersion : 0.000739229 seconds
Update interval : 64.4 seconds
Leap status     : Normal`,
	})
	res = snc.RunCheck("check_ntp_offset", []string{"source=chronyc"})
	assert.Equalf(t, CheckExitOK, res.State, "state OK")
	assert.Equalf(t, "OK - offset -19.2ms from test.ntp |'offset'=-19.212097ms;-50:50;-100:100 'stratum'=3;;;0",
		string(res.BuildPluginOutput()), "output matches")

	// mock freshly initialized chrony
	MockSystemUtilities(t, map[string]string{
		"chronyc": `Reference ID    : 00000000 ()
Stratum         : 0
Ref time (UTC)  : Thu Jan 01 00:00:00 1970
System time     : 0.000000000 seconds fast of NTP time
Last offset     : +0.000000000 seconds
RMS offset      : 0.000000000 seconds
Frequency       : 1.501 ppm slow
Residual freq   : +0.000 ppm
Skew            : 0.000 ppm
Root delay      : 1.000000000 seconds
Root dispersion : 1.000000000 seconds
Update interval : 0.0 seconds
Leap status     : Not synchronised`,
	})
	res = snc.RunCheck("check_ntp_offset", []string{"source=chronyc"})
	assert.Equalf(t, CheckExitCritical, res.State, "state Critical")
	assert.Equalf(t, "CRITICAL - chronyc: Leap status : Not synchronised",
		string(res.BuildPluginOutput()), "output matches")

	StopTestAgent(t, snc)
}

func TestCheckNTPOffsetNTPQ(t *testing.T) {
	snc := StartTestAgent(t, "")

	// mock ntpq command from output of: ntpq -p
	tmpPath := MockSystemUtilities(t, map[string]string{
		"ntpq": `     remote                                   refid      st t when poll reach   delay   offset   jitter
=======================================================================================================
 2.rhel.pool.ntp.org                     .POOL.          16 p    -  256    0   0.0000   0.0000   0.0001
+stratum2-3.ntp.techfak.net              129.70.137.82    2 u   48   64  377  32.0115  -1.5726   0.8925
-formularfetischisten.de                 131.188.3.223    2 u   51   64  377  26.8298   0.1340   0.8812
+185.13.148.71                           79.133.44.146    2 u   45   64  377  33.9011  -1.7828   0.7742
*ntp3.sack.dev                           129.69.1.153     2 u   47   64  377  21.6749  -1.1641   0.8209
`,
	})
	defer os.RemoveAll(tmpPath)
	res := snc.RunCheck("check_ntp_offset", []string{"source=ntpq"})
	assert.Equalf(t, CheckExitOK, res.State, "state OK")
	assert.Equalf(t, "OK - offset -1.164ms from ntp3.sack.dev (129.69.1.153) |'offset'=-1.1641ms;-50:50;-100:100 'stratum'=2;;;0 'jitter'=0.8209ms;;;0",
		string(res.BuildPluginOutput()), "output matches")

	// mock critical response
	MockSystemUtilities(t, map[string]string{
		"ntpq": `     remote                                   refid      st t when poll reach   delay   offset   jitter
=======================================================================================================
 2.rhel.pool.ntp.org                     .POOL.          16 p    -  256    0   0.0000   0.0000   0.0001
+stratum2-3.ntp.techfak.net              129.70.137.82    2 u   48   64  377  32.0115  -1.5726   0.8925
-formularfetischisten.de                 131.188.3.223    2 u   51   64  377  26.8298   0.1340   0.8812
+185.13.148.71                           79.133.44.146    2 u   45   64  377  33.9011  -1.7828   0.7742
*ntp3.sack.dev                           129.69.1.153     2 u   47   64  377  21.6749  -101.1641 0.8209
`,
	})
	res = snc.RunCheck("check_ntp_offset", []string{"source=ntpq"})
	assert.Equalf(t, CheckExitCritical, res.State, "state Critical")
	assert.Equalf(t, "CRITICAL - offset -101ms from ntp3.sack.dev (129.69.1.153) |'offset'=-101.1641ms;-50:50;-100:100 'stratum'=2;;;0 'jitter'=0.8209ms;;;0",
		string(res.BuildPluginOutput()), "output matches")

	// mock unknown response
	MockSystemUtilities(t, map[string]string{
		"ntpq": `     remote                                   refid      st t when poll reach   delay   offset   jitter
=======================================================================================================
 2.rhel.pool.ntp.org                     .POOL.          16 p    -  256    0   0.0000   0.0000   0.0001
 stratum2-2.ntp.techfak.net              129.70.137.82    2 u    -   64    1  34.1426  -1.3754   0.9159
 mail.gunnarhofmann.de                   192.53.103.103   2 u    1   64    1  27.1094  -0.8311   0.8566
 ns.gunnarhofmann.de                     205.46.178.169   2 u    1   64    1  26.9568  -0.6989   1.4830
 fa.gnudb.org                            192.38.7.240     2 u    1   64    1  27.0367  -0.3475   0.8936
 stratum2-4.ntp.techfak.net              .INIT.          16 u    -   64    0   0.0000   0.0000   0.0001
 timeserver.pub                          .INIT.          16 u    -   64    0   0.0000   0.0000   0.0001
 ntp01.pingless.com                      .INIT.          16 u    -   64    0   0.0000   0.0000   0.0001
 time.cloudflare.com                     .INIT.          16 u    -   64    0   0.0000   0.0000   0.0001`,
	})
	defer os.RemoveAll(tmpPath)
	res = snc.RunCheck("check_ntp_offset", []string{"source=ntpq"})
	assert.Equalf(t, CheckExitCritical, res.State, "state Critical")
	assert.Containsf(t, string(res.BuildPluginOutput()), "CRITICAL - ntpq did not return any usable server", "output matches")

	// single peer
	MockSystemUtilities(t, map[string]string{
		"ntpq": `     remote                                   refid      st t when poll reach   delay   offset   jitter
=======================================================================================================
+ntp.company                            123.123.123.123    2 u   13   64  377   0.3956 -415.442   0.6233`,
	})
	defer os.RemoveAll(tmpPath)
	res = snc.RunCheck("check_ntp_offset", []string{"source=ntpq"})
	assert.Equalf(t, CheckExitCritical, res.State, "state Critical")
	assert.Equalf(t, "CRITICAL - offset -415ms from ntp.company (123.123.123.123) |'offset'=-415.442ms;-50:50;-100:100 'stratum'=2;;;0 'jitter'=0.6233ms;;;0",
		string(res.BuildPluginOutput()), "output matches")

	StopTestAgent(t, snc)
}

func TestCheckNTPOffsetW32TM(t *testing.T) {
	snc := StartTestAgent(t, "")

	// mock ntpq command from output of: w32tm.exe /query /status /verbose
	tmpPath := MockSystemUtilities(t, map[string]string{
		"w32tm.exe": `Leap Indicator: 0(no warning)
Stratum: 4 (secondary reference - syncd by (S)NTP)
Precision: -6 (15.625ms per tick)
Root Delay: 0.0385101s
Root Dispersion: 0.0281350s
ReferenceId: 0x14653909 (source IP:  20.101.57.9)
Last Successful Sync Time: 12/20/2023 9:30:13 AM
Source: time.windows.com,0x8
Poll Interval: 10 (1024s)

Phase Offset: 0.0061517s
ClockRate: 0.0156215s
State Machine: 2 (Sync)
Time Source Flags: 0 (None)
Server Role: 0 (None)
Last Sync Error: 0 (The command completed successfully.)
Time since Last Good Sync Time: 339.9333552s`,
	})
	defer os.RemoveAll(tmpPath)
	res := snc.RunCheck("check_ntp_offset", []string{"source=w32tm"})
	assert.Equalf(t, CheckExitOK, res.State, "state OK")
	assert.Equalf(t, "OK - offset 6.152ms from time.windows.com |'offset'=6.1517ms;-50:50;-100:100 'stratum'=4;;;0",
		string(res.BuildPluginOutput()), "output matches")

	// mock critical response
	MockSystemUtilities(t, map[string]string{
		"w32tm.exe": `Leap Indicator: 0(no warning)
Stratum: 4 (secondary reference - syncd by (S)NTP)
Precision: -6 (15.625ms per tick)
Root Delay: 0.0385101s
Root Dispersion: 1.9281350s
ReferenceId: 0x14653909 (source IP:  20.101.57.9)
Last Successful Sync Time: 12/20/2023 9:30:13 AM
Source: time.windows.com,0x8
Poll Interval: 10 (1024s)

Phase Offset: 0.3061517s
ClockRate: 0.0156215s
State Machine: 2 (Sync)
Time Source Flags: 0 (None)
Server Role: 0 (None)
Last Sync Error: 0 (The command completed successfully.)
Time since Last Good Sync Time: 339.9333552s`,
	})
	res = snc.RunCheck("check_ntp_offset", []string{"source=w32tm"})
	assert.Equalf(t, CheckExitCritical, res.State, "state Critical")
	assert.Equalf(t, "CRITICAL - offset 306ms from time.windows.com |'offset'=306.1517ms;-50:50;-100:100 'stratum'=4;;;0",
		string(res.BuildPluginOutput()), "output matches")

	// mock unknown response from disabled service
	MockSystemUtilities(t, map[string]string{
		"w32tm.exe": `The following error occurred: The service has not been started. (0x80070426)`,
	})
	res = snc.RunCheck("check_ntp_offset", []string{"source=w32tm"})
	assert.Equalf(t, CheckExitCritical, res.State, "state Critical")
	assert.Equalf(t, "CRITICAL - cannot parse offset from w32tm: The following error occurred: The service has not been started. (0x80070426)\n",
		string(res.BuildPluginOutput()), "output matches")

	// mock unknown response from no network
	MockSystemUtilities(t, map[string]string{
		"w32tm.exe": `Leap Indicator: 0(no warning)
Stratum: 4 (secondary reference - syncd by (S)NTP)
Precision: -6 (15.625ms per tick)
Root Delay: 0.0379833s
Root Dispersion: 7.8604008s
ReferenceId: 0x28779426 (source IP:  40.119.148.38)
Last Successful Sync Time: 12/20/2023 10:26:32 AM
Source: time.windows.com,0x8
Poll Interval: 6 (64s)

Phase Offset: 0.0000002s
ClockRate: 0.0156250s
State Machine: 1 (Hold)
Time Source Flags: 0 (None)
Server Role: 0 (None)
Last Sync Error: 0 (The command completed successfully.)
Time since Last Good Sync Time: 46.3012345s`,
	})
	res = snc.RunCheck("check_ntp_offset", []string{"source=w32tm"})
	assert.Equalf(t, CheckExitCritical, res.State, "state Critical")
	assert.Equalf(t, "CRITICAL - w32tm.exe: State Machine: 1 (Hold)",
		string(res.BuildPluginOutput()), "output matches")

	StopTestAgent(t, snc)
}

func TestCheckNTPOffsetOSX(t *testing.T) {
	snc := StartTestAgent(t, "")

	// mock sntp and systemsetup commands
	tmpPath := MockSystemUtilities(t, map[string]string{
		"systemsetup": `Network Time: On
Network Time Server: time.euro.apple.com`,
		"sntp": `selected:
		sntp_exchange {
				result: 0 (Success)
				header: 24 (li:0 vn:4 mode:4)
			   stratum: 02 (2)
				  poll: 00 (1)
			 precision: FFFFFFE7 (2.980232e-08)
				 delay: 0000.0396 (0.014007568)
			dispersion: 0000.0003 (0.000045776)
				   ref: ED11CC5F (95.204.17.237)
				 t_ref: E92D4E4B.2335A41E (3912060491.137537247)
					t1: E92D4E62.B7E5B856 (3912060514.718348999)
					t2: E92D4E5A.CA878361 (3912060506.791130267)
					t3: E92D4E5A.CA896F1A (3912060506.791159576)
					t4: E92D4E62.B817EBAF (3912060514.719114999)
				offset: FFFFFFFFFFFFFFF8.1289A73B00000000 (-0.007587078)
				 delay: 0000000000000000.003047A000000000 (0.000736691)
				  mean: 00000000E92D4E5A.CA88793D80000000 (3912060506.791144848)
				 error: 0000000000000000.01CE000000000000 (0.007049561)
				  addr: 10.1.1.1
		}`,
	})
	defer os.RemoveAll(tmpPath)
	res := snc.RunCheck("check_ntp_offset", []string{"source=osx"})
	assert.Equalf(t, CheckExitOK, res.State, "state OK")
	assert.Equalf(t, "OK - offset -7.587ms from time.euro.apple.com (10.1.1.1) |'offset'=-7.587078ms;-50:50;-100:100 'stratum'=2;;;0",
		string(res.BuildPluginOutput()), "output matches")

	// mock unknown result
	MockSystemUtilities(t, map[string]string{
		"systemsetup": `Network Time: Off
Network Time Server: time.euro.apple.com`,
		"sntp": ``,
	})
	res = snc.RunCheck("check_ntp_offset", []string{"source=osx"})
	assert.Equalf(t, CheckExitCritical, res.State, "state Critical")
	assert.Containsf(t, string(res.BuildPluginOutput()), "CRITICAL - systemsetup -getusingnetworktime: Network Time: Off", "output matches")

	// mock unknown result
	MockSystemUtilities(t, map[string]string{
		"systemsetup": `Network Time: On
Network Time Server: time.euro.apple.com`,
		"sntp": `sntp_exchange {
			result: 6 (Timeout)
			header: 00 (li:0 vn:0 mode:0)
		   stratum: 00 (0)
			  poll: 00 (1)
		 precision: 00 (1.000000e+00)
			 delay: 0000.0000 (0.000000000)
		dispersion: 0000.0000 (0.000000000)
			   ref: 00000000 ("    ")
			 t_ref: 00000000.00000000 (0.000000000)
				t1: E92D4E77.853A2DF9 (3912060535.520418999)
				t2: 00000000.00000000 (0.000000000)
				t3: 00000000.00000000 (0.000000000)
				t4: 00000000.00000000 (0.000000000)
			offset: FFFFFFFF8B6958C4.3D62E90380000000 (-1956030267.760209560)
			 delay: FFFFFFFF16D2B188.7AC5D20700000000 (-3912060535.520419121)
			  mean: 0000000000000000.0000000000000000 (0.000000000)
			 error: 0000000000000000.0000000000000000 (0.000000000)
			  addr: 2a01:b740:a30:4000::1f2
	}`,
	})
	res = snc.RunCheck("check_ntp_offset", []string{"source=osx"})
	assert.Equalf(t, CheckExitCritical, res.State, "state Critical")
	assert.Equalf(t, "CRITICAL - sntp: result: 6 (Timeout)",
		string(res.BuildPluginOutput()), "output matches")

	StopTestAgent(t, snc)
}
