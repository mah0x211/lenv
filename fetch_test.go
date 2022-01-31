package main

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseLuaRocksVers(t *testing.T) {
	exp := []string{"luarocks-3.5.0.tar.gz", "luarocks-3.4.0.tar.gz", "luarocks-3.3.1.tar.gz", "luarocks-3.3.0.tar.gz", "luarocks-3.2.1.tar.gz", "luarocks-3.2.0.tar.gz", "luarocks-3.1.3.tar.gz", "luarocks-3.1.2.tar.gz", "luarocks-3.1.1.tar.gz", "luarocks-3.1.0.tar.gz", "luarocks-3.0.4.tar.gz", "luarocks-3.0.3.tar.gz", "luarocks-3.0.2.tar.gz", "luarocks-3.0.1.tar.gz", "luarocks-3.0.1-rc2.tar.gz", "luarocks-3.0.1-rc1.tar.gz", "luarocks-3.0.0.tar.gz", "luarocks-3.0.0-rc2.tar.gz", "luarocks-3.0.0-rc1.tar.gz", "luarocks-3.0.0beta2.tar.gz", "luarocks-3.0.0beta1.tar.gz", "luarocks-2.4.4.tar.gz", "luarocks-2.4.3.tar.gz", "luarocks-2.4.2.tar.gz", "luarocks-2.4.1.tar.gz", "luarocks-2.4.0.tar.gz", "luarocks-2.3.0.tar.gz", "luarocks-2.3.0-rc2.tar.gz", "luarocks-2.3.0-rc1.tar.gz", "luarocks-2.2.3-rc2.tar.gz", "luarocks-2.2.3-rc1.tar.gz", "luarocks-2.2.2.tar.gz", "luarocks-2.2.1.tar.gz", "luarocks-2.2.0.tar.gz", "luarocks-2.2.0beta1.tar.gz", "luarocks-2.1.2.tar.gz", "luarocks-2.1.1.tar.gz", "luarocks-2.1.0.tar.gz", "luarocks-2.1.0-rc3.tar.gz", "luarocks-2.1.0-rc2.tar.gz", "luarocks-2.1.0-rc1.tar.gz", "luarocks-2.0.13.tar.gz", "luarocks-2.0.12.tar.gz", "luarocks-2.0.11.tar.gz", "luarocks-2.0.10.tar.gz", "luarocks-2.0.9.1.tar.gz", "luarocks-2.0.9.tar.gz", "luarocks-2.0.9-rc2.tar.gz", "luarocks-2.0.9-rc1.tar.gz", "luarocks-2.0.8.tar.gz", "luarocks-2.0.8-rc2.tar.gz", "luarocks-2.0.8-rc1.tar.gz", "luarocks-2.0.7.1.tar.gz", "luarocks-2.0.7.tar.gz", "luarocks-2.0.6.tar.gz", "luarocks-2.0.6-rc1.tar.gz", "luarocks-2.0.5.tar.gz", "luarocks-2.0.5-rc1.tar.gz", "luarocks-2.0.4.1.tar.gz", "luarocks-2.0.4.tar.gz", "luarocks-2.0.4-rc3.tar.gz", "luarocks-2.0.4-rc2.tar.gz", "luarocks-2.0.4-rc1.tar.gz", "luarocks-2.0.3.tar.gz", "luarocks-2.0.3-rc2.tar.gz", "luarocks-2.0.3-rc1.tar.gz", "luarocks-2.0.2.tar.gz", "luarocks-2.0.1.tar.gz", "luarocks-2.0.tar.gz", "luarocks-1.0.1.tar.gz", "luarocks-1.0.tar.gz", "luarocks-0.6.0.2.tar.gz", "luarocks-0.5.2.tar.gz", "luarocks-0.5.1.tar.gz", "luarocks-0.5.tar.gz", "luarocks-0.4.3.tar.gz", "luarocks-0.4.2.tar.gz", "luarocks-0.4.1.tar.gz", "luarocks-0.4.tar.gz", "luarocks-0.3.2.tar.gz", "luarocks-0.3.1.tar.gz", "luarocks-0.3.tar.gz", "luarocks-0.2.tar.gz", "luarocks-0.1.tar.gz"}
	names := []string{}

	for ver, item := range parseLuaRocksVers([]byte(LuaRocksHTML)) {
		names = append(names, item.Name)
		assert.Equal(t, ver, item.Ver)
	}
	sort.Strings(exp)
	sort.Strings(names)
	assert.Equal(t, exp, names)
}

func Test_parseLuaJitVers(t *testing.T) {
	exp := []string{"LuaJIT-2.1.0-beta3.tar.gz", "LuaJIT-2.0.5.tar.gz", "LuaJIT-1.1.8.tar.gz", "LuaJIT-1.0.3.tar.gz"}
	names := []string{}

	for ver, item := range parseLuaJitVers([]byte(LuaJitHTML)) {
		names = append(names, item.Name)
		assert.Equal(t, ver, item.Ver)
	}
	sort.Strings(exp)
	sort.Strings(names)
	assert.Equal(t, exp, names)
}

func Test_parseLuaVers(t *testing.T) {
	exp := []string{"lua-5.4.4.tar.gz", "lua-5.4.3.tar.gz", "lua-5.4.2.tar.gz", "lua-5.4.1.tar.gz", "lua-5.4.0.tar.gz", "lua-5.3.6.tar.gz", "lua-5.3.5.tar.gz", "lua-5.3.4.tar.gz", "lua-5.3.3.tar.gz", "lua-5.3.2.tar.gz", "lua-5.3.1.tar.gz", "lua-5.3.0.tar.gz", "lua-5.2.4.tar.gz", "lua-5.2.3.tar.gz", "lua-5.2.2.tar.gz", "lua-5.2.1.tar.gz", "lua-5.2.0.tar.gz", "lua-5.1.5.tar.gz", "lua-5.1.4.tar.gz", "lua-5.1.3.tar.gz", "lua-5.1.2.tar.gz", "lua-5.1.1.tar.gz", "lua-5.1.tar.gz", "lua-5.0.3.tar.gz", "lua-5.0.2.tar.gz", "lua-5.0.1.tar.gz", "lua-5.0.tar.gz", "lua-4.0.1.tar.gz", "lua-4.0.tar.gz", "lua-3.2.2.tar.gz", "lua-3.2.1.tar.gz", "lua-3.2.tar.gz", "lua-3.1.tar.gz", "lua-3.0.tar.gz", "lua-2.5.tar.gz", "lua-2.4.tar.gz", "lua-2.2.tar.gz", "lua-2.1.tar.gz", "lua-1.1.tar.gz", "lua-1.0.tar.gz"}
	names := []string{}

	for ver, item := range parseLuaVers([]byte(LuaHTML)) {
		names = append(names, item.Name)
		assert.Equal(t, ver, item.Ver)
	}
	sort.Strings(names)
	sort.Strings(exp)
	assert.Equal(t, exp, names)
}

const LuaHTML = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">
<HTML>
<HEAD>
<TITLE>Lua: download area</TITLE>
<LINK REL="stylesheet" TYPE="text/css" HREF="../lua.css">
<LINK REL="stylesheet" TYPE="text/css" HREF="../ftp.css">
<META HTTP-EQUIV="content-type" CONTENT="text/html; charset=iso-8859-1">
</HEAD>

<BODY>

<H1>
<A HREF="../home.html"><IMG SRC="../images/logo.gif" ALT="Lua"></A>
Download area
</H1>

<DIV CLASS="menubar">
<A HREF="#source">source</A>
&middot;
<A HREF="#manuals">manuals</A>
&middot;
<A HREF="../license.html">license</A>
&middot;
<A HREF="../versions.html">versions</A>
&middot;
<A HREF="../work/">work area</A>
</DIV>

<P>
This repository contains
the <A HREF="#source">source code</A> and
the <A HREF="#manuals">reference manuals</A>
for all <A HREF="../versions.html">released versions</A> of Lua.

<P>
If you want to build early versions of Lua using modern compilers, get the
<A HREF="lua-all.tar.gz">lua-all</A> package.

<P>
If you have arrived here by accident,
<A HREF="../">start here</A>.

<P>
If you are looking for work versions, release candidates, and other pre-releases of Lua, check out the
<A HREF="../work/">work area</A>.

<P>
All files are distributed under this
<A HREF="../license.html">license</A>.
Check their checksums to confirm the integrity of the packages.

<H2><A NAME="source">Source code</A></H2>
<TABLE>
<TR>
<TH>filename</TH>
<TH>date</TH>
<TH>size</TH>
<TH>checksum (sha256)</TH>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.4.4.tar.gz">lua-5.4.4.tar.gz</A></TD>
<TD CLASS="date">2022-01-13</TD>
<TD CLASS="size">360876</TD>
<TD CLASS="sum">164c7849653b80ae67bec4b7473b884bf5cc8d2dca05653475ec2ed27b9ebf61</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.4.3.tar.gz">lua-5.4.3.tar.gz</A></TD>
<TD CLASS="date">2021-03-15</TD>
<TD CLASS="size">358216</TD>
<TD CLASS="sum">f8612276169e3bfcbcfb8f226195bfc6e466fe13042f1076cbde92b7ec96bbfb</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.4.2.tar.gz">lua-5.4.2.tar.gz</A></TD>
<TD CLASS="date">2020-11-13</TD>
<TD CLASS="size">353472</TD>
<TD CLASS="sum">11570d97e9d7303c0a59567ed1ac7c648340cd0db10d5fd594c09223ef2f524f</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.4.1.tar.gz">lua-5.4.1.tar.gz</A></TD>
<TD CLASS="date">2020-09-30</TD>
<TD CLASS="size">353965</TD>
<TD CLASS="sum">4ba786c3705eb9db6567af29c91a01b81f1c0ac3124fdbf6cd94bdd9e53cca7d</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.4.0.tar.gz">lua-5.4.0.tar.gz</A></TD>
<TD CLASS="date">2020-06-18</TD>
<TD CLASS="size">349308</TD>
<TD CLASS="sum">eac0836eb7219e421a96b7ee3692b93f0629e4cdb0c788432e3d10ce9ed47e28</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.6.tar.gz">lua-5.3.6.tar.gz</A></TD>
<TD CLASS="date">2020-09-14</TD>
<TD CLASS="size">303770</TD>
<TD CLASS="sum">fc5fd69bb8736323f026672b1b7235da613d7177e72558893a0bdcd320466d60</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.5.tar.gz">lua-5.3.5.tar.gz</A></TD>
<TD CLASS="date">2018-06-26</TD>
<TD CLASS="size">303543</TD>
<TD CLASS="sum">0c2eed3f960446e1a3e4b9a1ca2f3ff893b6ce41942cf54d5dd59ab4b3b058ac</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.4.tar.gz">lua-5.3.4.tar.gz</A></TD>
<TD CLASS="date">2017-01-12</TD>
<TD CLASS="size">303586</TD>
<TD CLASS="sum">f681aa518233bc407e23acf0f5887c884f17436f000d453b2491a9f11a52400c</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.3.tar.gz">lua-5.3.3.tar.gz</A></TD>
<TD CLASS="date">2016-05-30</TD>
<TD CLASS="size">294290</TD>
<TD CLASS="sum">5113c06884f7de453ce57702abaac1d618307f33f6789fa870e87a59d772aca2</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.2.tar.gz">lua-5.3.2.tar.gz</A></TD>
<TD CLASS="date">2015-11-25</TD>
<TD CLASS="size">288235</TD>
<TD CLASS="sum">c740c7bb23a936944e1cc63b7c3c5351a8976d7867c5252c8854f7b2af9da68f</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.1.tar.gz">lua-5.3.1.tar.gz</A></TD>
<TD CLASS="date">2015-06-10</TD>
<TD CLASS="size">282401</TD>
<TD CLASS="sum">072767aad6cc2e62044a66e8562f51770d941e972dc1e4068ba719cd8bffac17</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.0.tar.gz">lua-5.3.0.tar.gz</A></TD>
<TD CLASS="date">2015-01-06</TD>
<TD CLASS="size">278045</TD>
<TD CLASS="sum">ae4a5eb2d660515eb191bfe3e061f2b8ffe94dce73d32cfd0de090ddcc0ddb01</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.2.4.tar.gz">lua-5.2.4.tar.gz</A></TD>
<TD CLASS="date">2015-02-26</TD>
<TD CLASS="size">252651</TD>
<TD CLASS="sum">b9e2e4aad6789b3b63a056d442f7b39f0ecfca3ae0f1fc0ae4e9614401b69f4b</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.2.3.tar.gz">lua-5.2.3.tar.gz</A></TD>
<TD CLASS="date">2013-11-11</TD>
<TD CLASS="size">251195</TD>
<TD CLASS="sum">13c2fb97961381f7d06d5b5cea55b743c163800896fd5c5e2356201d3619002d</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.2.2.tar.gz">lua-5.2.2.tar.gz</A></TD>
<TD CLASS="date">2013-03-21</TD>
<TD CLASS="size">251713</TD>
<TD CLASS="sum">3fd67de3f5ed133bf312906082fa524545c6b9e1b952e8215ffbd27113f49f00</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.2.1.tar.gz">lua-5.2.1.tar.gz</A></TD>
<TD CLASS="date">2012-06-08</TD>
<TD CLASS="size">249882</TD>
<TD CLASS="sum">64304da87976133196f9e4c15250b70f444467b6ed80d7cfd7b3b982b5177be5</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.2.0.tar.gz">lua-5.2.0.tar.gz</A></TD>
<TD CLASS="date">2011-12-12</TD>
<TD CLASS="size">246377</TD>
<TD CLASS="sum">cabe379465aa8e388988073d59b69e76ba0025429d2c1da80821a252cdf6be0d</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.1.5.tar.gz">lua-5.1.5.tar.gz</A></TD>
<TD CLASS="date">2012-02-13</TD>
<TD CLASS="size">221213</TD>
<TD CLASS="sum">2640fc56a795f29d28ef15e13c34a47e223960b0240e8cb0a82d9b0738695333</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.1.4.tar.gz">lua-5.1.4.tar.gz</A></TD>
<TD CLASS="date">2008-08-18</TD>
<TD CLASS="size">216679</TD>
<TD CLASS="sum">b038e225eaf2a5b57c9bcc35cd13aa8c6c8288ef493d52970c9545074098af3a</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.1.3.tar.gz">lua-5.1.3.tar.gz</A></TD>
<TD CLASS="date">2008-01-21</TD>
<TD CLASS="size">215817</TD>
<TD CLASS="sum">6b5df2edaa5e02bf1a2d85e1442b2e329493b30b0c0780f77199d24f087d296d</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.1.2.tar.gz">lua-5.1.2.tar.gz</A></TD>
<TD CLASS="date">2007-03-29</TD>
<TD CLASS="size">214134</TD>
<TD CLASS="sum">5cf098c6fe68d3d2d9221904f1017ff0286e4a9cc166a1452a456df9b88b3d9e</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.1.1.tar.gz">lua-5.1.1.tar.gz</A></TD>
<TD CLASS="date">2006-06-07</TD>
<TD CLASS="size">207810</TD>
<TD CLASS="sum">c5daeed0a75d8e4dd2328b7c7a69888247868154acbda69110e97d4a6e17d1f0</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.1.tar.gz">lua-5.1.tar.gz</A></TD>
<TD CLASS="date">2006-02-20</TD>
<TD CLASS="size">206877</TD>
<TD CLASS="sum">7f5bb9061eb3b9ba1e406a5aa68001a66cb82bac95748839dc02dd10048472c1</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.0.3.tar.gz">lua-5.0.3.tar.gz</A></TD>
<TD CLASS="date">2006-06-19</TD>
<TD CLASS="size">191384</TD>
<TD CLASS="sum">1193a61b0e08acaa6eee0eecf29709179ee49c71baebc59b682a25c3b5a45671</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.0.2.tar.gz">lua-5.0.2.tar.gz</A></TD>
<TD CLASS="date">2004-03-17</TD>
<TD CLASS="size">190442</TD>
<TD CLASS="sum">a6c85d85f912e1c321723084389d63dee7660b81b8292452b190ea7190dd73bc</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.0.1.tar.gz">lua-5.0.1.tar.gz</A></TD>
<TD CLASS="date">2003-11-25</TD>
<TD CLASS="size">193978</TD>
<TD CLASS="sum">7a09d0e70dcaff7feae97cf9c154da05b1e5b92eaea2df7150b54bcaf8f3b9c6</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.0.tar.gz">lua-5.0.tar.gz</A></TD>
<TD CLASS="date">2003-04-11</TD>
<TD CLASS="size">187287</TD>
<TD CLASS="sum">4a23b3bcb812538c653033cd39fe9c9bd8030286b945c56eff280d452e4e244e</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-4.0.1.tar.gz">lua-4.0.1.tar.gz</A></TD>
<TD CLASS="date">2002-07-04</TD>
<TD CLASS="size">158426</TD>
<TD CLASS="sum">df746e149cf6939e90009d2e540eee918d585b4d1bc6d68b19316a050d484d2a</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-4.0.tar.gz">lua-4.0.tar.gz</A></TD>
<TD CLASS="date">2000-11-06</TD>
<TD CLASS="size">157102</TD>
<TD CLASS="sum">b476abc737bf82781cb215e59a259bf23adbdc82425907b51f37ba29c0e71337</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-3.2.2.tar.gz">lua-3.2.2.tar.gz</A></TD>
<TD CLASS="date">2000-02-22</TD>
<TD CLASS="size">127768</TD>
<TD CLASS="sum">4e04059f43acdcde5f7fd491c731df9279dac87d288a08c6eaeb31760c9876e0</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-3.2.1.tar.gz">lua-3.2.1.tar.gz</A></TD>
<TD CLASS="date">1999-11-25</TD>
<TD CLASS="size">127644</TD>
<TD CLASS="sum">00d156f5b0c99bf46ed2fc2471b2a4bc5762aee83a809ab87bd7ec9ccc6220ea</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-3.2.tar.gz">lua-3.2.tar.gz</A></TD>
<TD CLASS="date">1999-07-08</TD>
<TD CLASS="size">128597</TD>
<TD CLASS="sum">bf8beabd41e65cbf8cb41c688eca0588fff81e1e5f67cb42bd370e1ecc585c33</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-3.1.tar.gz">lua-3.1.tar.gz</A></TD>
<TD CLASS="date">1998-07-12</TD>
<TD CLASS="size">114186</TD>
<TD CLASS="sum">797ceabe28df53ce6dd6bfb4397d6e0b2e349b3bd0a1883c9c8ae0212d8cd7ed</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-3.0.tar.gz">lua-3.0.tar.gz</A></TD>
<TD CLASS="date">1997-07-01</TD>
<TD CLASS="size">99921</TD>
<TD CLASS="sum">a3ae0fb0ffc6b365d187fea0d121772fc2a8810853b14b7e48ed31eae8494215</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-2.5.tar.gz">lua-2.5.tar.gz</A></TD>
<TD CLASS="date">1996-11-21</TD>
<TD CLASS="size">185786</TD>
<TD CLASS="sum">8ba8cc089b6ee7b5ed7fe9e0764756838943a1e52dde8bbfa70b758556056cd5</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-2.4.tar.gz">lua-2.4.tar.gz</A></TD>
<TD CLASS="date">1996-05-17</TD>
<TD CLASS="size">132500</TD>
<TD CLASS="sum">38ee28cd5ec5c0b75f31d65c12f587f192174c8b7fcca3decd08cdb35af282e3</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-2.2.tar.gz">lua-2.2.tar.gz</A></TD>
<TD CLASS="date">1995-11-28</TD>
<TD CLASS="size">108261</TD>
<TD CLASS="sum">a13cc27035b8d16bd9d29f5a1eb4403476deecd149c1d3979ad9d2ad015994f3</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-2.1.tar.gz">lua-2.1.tar.gz</A></TD>
<TD CLASS="date">1995-09-13</TD>
<TD CLASS="size">125334</TD>
<TD CLASS="sum">d65ee99f8e63d12fa952425671854add39c1233e4a9c6669ec552b16279df664</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-1.1.tar.gz">lua-1.1.tar.gz</A></TD>
<TD CLASS="date">1995-02-02</TD>
<TD CLASS="size">158285</TD>
<TD CLASS="sum">e44506793fee609202d7f42e529afbc3c9a541e0fc715b3834d758f0ae41c38f</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-1.0.tar.gz">lua-1.0.tar.gz</A></TD>
<TD CLASS="date">2003-10-10</TD>
<TD CLASS="size">33149</TD>
<TD CLASS="sum">d8ee432490a4679a4edb0abbef2d3b7b0f1fc29a39398988c4bb5eeb8d2180a6</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-all.tar.gz">lua-all.tar.gz</A></TD>
<TD CLASS="date">2022-01-26</TD>
<TD CLASS="size">5142787</TD>
<TD CLASS="sum">a1c1d96a1e1968ad598b708e57aab720f60247d333c4d9a46f99a2d2bd32209d</TD>
</TR>

</TABLE>

<H2><A NAME="manuals">Manuals</A></H2>
<TABLE>
<TR>
<TH>filename</TH>
<TH>date</TH>
<TH>size</TH>
<TH>checksum (sha256)</TH>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-5.4.tar.gz">refman-5.4.tar.gz</A></TD>
<TD CLASS="date">2021-06-17</TD>
<TD CLASS="size">124407</TD>
<TD CLASS="sum">5321b0045d54cbd711a6a0da3bb12f9380e8d4081b91d726171551f27d1245eb</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-5.3.tar.gz">refman-5.3.tar.gz</A></TD>
<TD CLASS="date">2020-09-25</TD>
<TD CLASS="size">110173</TD>
<TD CLASS="sum">74fc89f552d6cbbcbd63edb2bf309f808e9c69fe1350eaaf254c4fecb13a0708</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-5.2.tar.gz">refman-5.2.tar.gz</A></TD>
<TD CLASS="date">2016-07-01</TD>
<TD CLASS="size">102134</TD>
<TD CLASS="sum">562da8f399ea55eb2c85bdfd2d93d90b7cc7ea7b041c41718d70ec9ed8d085c4</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-5.1.tar.gz">refman-5.1.tar.gz</A></TD>
<TD CLASS="date">2016-07-01</TD>
<TD CLASS="size">86641</TD>
<TD CLASS="sum">7d73f55384a862f0f032feabfa8de96203ba03fcf4179491bad232d2ee0f3b3e</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-5.0.ps.gz">refman-5.0.ps.gz</A></TD>
<TD CLASS="date">2003-11-25</TD>
<TD CLASS="size">224198</TD>
<TD CLASS="sum">d3a73f9435fec6582f88c2efdb0d2ede4d2824ff595cdcc766c44fb08684785d</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-5.0.pdf">refman-5.0.pdf</A></TD>
<TD CLASS="date">2003-11-25</TD>
<TD CLASS="size">474818</TD>
<TD CLASS="sum">07cc73d3e9a818067eac1f3cf904ac26834c36f46bd2125d2325358c5a380ad5</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-4.0.ps.gz">refman-4.0.ps.gz</A></TD>
<TD CLASS="date">2000-11-06</TD>
<TD CLASS="size">156876</TD>
<TD CLASS="sum">81124c666d219c8f6c144012f992d74bef23e052fdd24b7c469250ef9d811e6c</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-4.0.pdf">refman-4.0.pdf</A></TD>
<TD CLASS="date">2000-11-07</TD>
<TD CLASS="size">446955</TD>
<TD CLASS="sum">c4dfc6046ba09a64f0062f65d86016408f0fcf5d9ae6c16dc24c51e333b304ee</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-3.2.ps.gz">refman-3.2.ps.gz</A></TD>
<TD CLASS="date">1999-07-02</TD>
<TD CLASS="size">136552</TD>
<TD CLASS="sum">401ecac840bdea0f3cf129a9b74eea349cd6258fb1edc2af1caf69d84b0cecfc</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-3.2.pdf">refman-3.2.pdf</A></TD>
<TD CLASS="date">1999-07-08</TD>
<TD CLASS="size">321348</TD>
<TD CLASS="sum">454dc05959e8b9c7a72fcb261643df9abc2028f0e394eb7b7a6ab9c6561344e6</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-3.1.ps.gz">refman-3.1.ps.gz</A></TD>
<TD CLASS="date">1998-07-01</TD>
<TD CLASS="size">129519</TD>
<TD CLASS="sum">7d0d0fa10eeb18965a2f06e1bfbe223a74fa9d2b4b68ad8951750f00eaaa896a</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-2.5.ps.gz">refman-2.5.ps.gz</A></TD>
<TD CLASS="date">1996-11-18</TD>
<TD CLASS="size">115860</TD>
<TD CLASS="sum">7f4951eb2eb3c1eceabf7571275c7f856d0f89342aa40e59b46e6c3c656508c0</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-2.4.ps.gz">refman-2.4.ps.gz</A></TD>
<TD CLASS="date">1996-05-14</TD>
<TD CLASS="size">69189</TD>
<TD CLASS="sum">8ae641b31cc70ed838ab4449825d2003c82cac5d399a3fa8d81a357c75cc5bd8</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-2.2.ps.gz">refman-2.2.ps.gz</A></TD>
<TD CLASS="date">1995-11-28</TD>
<TD CLASS="size">56724</TD>
<TD CLASS="sum">9c76439231a444c96bacccc378b06f09465324f9d38ce406376c1d2c5a7723ca</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-2.1.ps.gz">refman-2.1.ps.gz</A></TD>
<TD CLASS="date">1995-02-08</TD>
<TD CLASS="size">55171</TD>
<TD CLASS="sum">b3ef5e53b302083c7e14d721339abfbc8c0f74d1244771df2962f697a7a41eb1</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-1.1.ps.gz">refman-1.1.ps.gz</A></TD>
<TD CLASS="date">1994-05-27</TD>
<TD CLASS="size">45131</TD>
<TD CLASS="sum">39c920fcefc7983c9c785509f90f0cbb6d1404f2543f0c925c5eb2459bec8144</TD>
</TR>

</TABLE>

<P CLASS="footer">
Last update:
Mon Jan 31 12:31:37 UTC 2022
</P>
<!--
Last change: use sha256 checksums
-->

</BODY>
</HTML>

`

const LuaJitHTML = `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">
<html>
<head>
<title>Download</title>
<meta http-equiv="Content-Type" content="text/html; charset=iso-8859-1">
<meta name="Copyright" content="Copyright (C) 2005-2021">
<meta name="Language" content="en">
<link rel="stylesheet" type="text/css" href="bluequad.css" media="screen">
<link rel="stylesheet" type="text/css" href="bluequad-print.css" media="print">
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1.5">
<script type="text/javascript">
var n=null,m=0;function toggleMenu(t){m|=t;if(m==t&&n!==null){n=!n;document.getElementById("nav").style.display=n?"block":"none";return true;}}
if("ontouchstart" in window && /iPhone|iPad|iPod|Android/i.test(navigator.userAgent)){var l=document.createElement("link");l.id="mcss";l.rel="stylesheet";l.type="text/css";l.href="bluequad-touch.css";l.media="all";document.getElementsByTagName("head")[0].appendChild(l);n=false;}
</script>
<style type="text/css">
table.download { width: 36em; }
table.downloadfix { width: 42.2em; }
tr.downhead td { font-weight: bold; }
td.downname { font-weight: bold; }
td.downtime { text-align: center; width: 7em; }
td.downsize { font-weight: bold; text-align: right; width: 5.2em; }
td.downsize a:hover, td.downname a:active { text-decoration: none; }
td.downfix { width: 5.2em; }
</style>
</head>
<body>
<div id="site" ontouchstart="toggleMenu(1);" onclick="toggleMenu(0);">
<a href="https://luajit.org"><span>Lua<span id="logo">JIT</span></span></a>
</div>
<div id="head">
<h1>Download</h1>
</div>
<div id="nav">
<ul><li>
<a href="index.html">Home</a>
</li><li>
<a href="luajit.html">LuaJIT</a>
<ul><li>
<a class="current" href="download.html">Download &#8681;</a>
</li><li>
<a href="install.html">Installation</a>
</li><li>
<a href="running.html">Running</a>
</li></ul>
</li><li>
<a href="extensions.html">Extensions</a>
<ul><li>
<a href="ext_ffi.html">FFI Library</a>
<ul><li>
<a href="ext_ffi_tutorial.html">FFI Tutorial</a>
</li><li>
<a href="ext_ffi_api.html">ffi.* API</a>
</li><li>
<a href="ext_ffi_semantics.html">FFI Semantics</a>
</li></ul>
</li><li>
<a href="ext_jit.html">jit.* Library</a>
</li><li>
<a href="ext_c_api.html">Lua/C API</a>
</li></ul>
</li><li>
<a href="status.html">Status</a>
</li><li>
<a href="faq.html">FAQ</a>
</li><li>
<a href="performance.html">Performance</a>
<ul><li>
<a href="performance_x86.html">on x86/x64</a>
</li><li>
<a href="performance_arm.html">on ARM</a>
</li><li>
<a href="performance_ppc.html">on PPC</a>
</li><li>
<a href="performance_e500.html">on PPC/e500</a>
</li><li>
<a href="performance_mips.html">on MIPS</a>
</li></ul>
</li><li>
<a href="http://wiki.luajit.org/">Wiki <span class="ext">&raquo;</span></a>
</li><li>
<a href="list.html">Mailing List</a>
</li><li>
<a href="sponsors.html">Sponsors</a>
</li></ul>
</div>
<div id="main">
<h2>Public git Repository</h2>
<p>
The public <a href="https://git-scm.com">git</a> repository contains
the current state of the actively developed branches. You can clone
it with the following command:
</p>
<pre class="code" style="margin-bottom: 0;">
git clone https://luajit.org/git/luajit-2.0.git
</pre>
<p style="font-size: 8pt; color: #808080; margin-left: 0.8em;">
Note: this is not a browsable resource, the repository is only accessible
with a git client. Despite the name of the repo, it contains all branches.</p>
<p>
This creates a new directory tree under <tt>luajit-2.0</tt>. Change to it,
switch to the branch you want to use and follow the usual build instructions.
Use <tt>git pull</tt> to fetch updates from the (read-only) public repository.
There's also an
<a href="https://luajit.org/git/luajit-2.0.git/feed.xml">RSS feed</a>
for the commit messages, a
<a href="https://repo.or.cz/w/luajit-2.0.git">browsable mirror</a>
and a
<a href="https://github.com/LuaJIT/LuaJIT">GitHub mirror</a>.
</p>
<p>
The git <tt>master</tt> branch will be phased out and is pinned to the
<tt>v2.0</tt> branch. Please follow the versioned branches <tt>v2.1</tt>
or <tt>v2.0</tt> instead.
</p>

<h2>Releases of Actively Developed Branches</h2>

<p>
Releases are only made occasionally. You're strongly encouraged to follow
the git branches.
</p>
<p>
Distros that require releases should do <i>regular</i> snapshots of a branch.
Do not attempt to cherry-pick or backport individual changes, no matter
how self-standing individual changes look (because they often are not).
</p>
<p>
Note: The tar.gz and zip files of each release have the same contents
&mdash; you only need to download one of them.
</p>

<table class="download">
<tr class="downhead">
<td class="downname">Filename</td>
<td class="downtime">Date</td>
<td class="downsize">.tar.gz</td>
<td class="downsize">.zip</td>
</tr>
<tr class="odd separate">
<td class="downname">LuaJIT-2.1.0-beta3</td><td class="downtime">2017-05-01</td><td class="downsize"><a href="download/LuaJIT-2.1.0-beta3.tar.gz">1001K &#9660;</a></td><td class="downsize"><a href="download/LuaJIT-2.1.0-beta3.zip">1130K &#9660;</a></td></tr>
<tr class="even">
<td class="downname">LuaJIT-2.0.5</td><td class="downtime">2017-05-01</td><td class="downsize"><a href="download/LuaJIT-2.0.5.tar.gz">830K &#9660;</a></td><td class="downsize"><a href="download/LuaJIT-2.0.5.zip">940K &#9660;</a></td></tr>
</table>

<h2>Historic Releases</h2>
<p>
This is a list of the archived packages of historic releases. Please consider
using the actively developed versions, unless you have special needs.
</p>

<table class="download">
<tr class="downhead">
<td class="downname">Filename</td>
<td class="downtime">Date</td>
<td class="downsize">.tar.gz</td>
<td class="downsize">.zip</td>
</tr>
<tr class="odd separate">
<td class="downname">LuaJIT-1.1.8</td><td class="downtime">2012-04-16</td><td class="downsize"><a href="download/LuaJIT-1.1.8.tar.gz">362K &#9660;</a></td><td class="downsize"><a href="download/LuaJIT-1.1.8.zip">423K &#9660;</a></td></tr>
<tr class="even">
<td class="downname">LuaJIT-1.0.3</td><td class="downtime">2005-09-08</td><td class="downsize"><a href="download/LuaJIT-1.0.3.tar.gz">301K &#9660;</a></td><td class="downsize">&nbsp;</td></tr>
</table>

<h2>SHA256 Checksums</h2>
<pre style="margin: 0;">
1ad2e34b111c802f9d0cdf019e986909123237a28c746b21295b63c9e785d9c3  LuaJIT-2.1.0-beta3.tar.gz
fcc4069bfaf909f762844d6404a8c5940591b00237ffce1781e887a0964140da  LuaJIT-2.1.0-beta3.zip
874b1f8297c697821f561f9b73b57ffd419ed8f4278c82e05b48806d30c1e979  LuaJIT-2.0.5.tar.gz
95f655cff930781619f0fbbab707fb8c05f406d9007a3505c8243e3dbbfacec8  LuaJIT-2.0.5.zip
42f095d0215d76c29b7b040ad52dddc1783ffc6e3021b8a831627973a8a32862  LuaJIT-1.1.8.tar.gz
d59751726e6d3f22a4b3cfb158d728b8feea6aa3b9ee30af610685bc28b0e6c8  LuaJIT-1.1.8.zip
e39204aad8d2a3f9ef74a4a515fedf3cce3f55ff247af8ae50f2a8cb01f257c3  LuaJIT-1.0.3.tar.gz
</pre>
<br class="flush">
</div>
<div id="foot">
<hr class="hide">
Copyright &copy; 2005-2021
<span class="noprint">
&middot;
<a href="contact.html">Contact &ndash; IMPRESSUM</a>
</span>
</div>
</body>
</html>

`

const LuaRocksHTML = `<!doctype html>
<html>
<head>
    <link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Open+Sans:400italic,400,700" type="text/css"/>
    <link rel="stylesheet" href="../style.css" type="text/css"/>
    <style>
    table {
        width: 100%;
    }

    td {
        padding: 5px;
        background-color: white;
        border: 1px solid #c1cce4;
    }
    </style>
</head>
<body>
    <div class="content">
        <div class="header">
            <div class="header_inner">
                <div class="user_panel">
                    <a href="https://github.com/keplerproject/luarocks/wiki/Download">Install</a>
                     &middot;
                    <a href="https://github.com/keplerproject/luarocks/wiki/Documentation">Docs</a>
                     &middot;
                    <a href="https://luarocks.org/login">Log In</a>
                     &middot;
                    <a href="https://luarocks.org/register">Register</a>
                </div>
                <a href="https://luarocks.org/">
                    <img alt="LuaRocks" class="text_logo" src="https://luarocks.org/static/header_luarocks_name.svg"/>
                </a>
                <form method="GET" action="https://luarocks.org/search" class="header_search">
                    <input type="text" placeholder="Search modules or uploaders..." name="q"/>
                </form>
            </div>
        </div>
        <div class="index_page">
            <div class="main_column">

                <h2>LuaRocks releases</h2>
                <table>

                    <!-- add new release here -->

                    <tr>
                        <td>
                            <a href="luarocks-3.5.0.tar.gz">luarocks-3.5.0.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.5.0.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.5.0-windows-32.zip">luarocks-3.5.0-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows 32-bit binary)
                        </td>
                        <td>
                            <a href="luarocks-3.5.0-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.5.0-windows-64.zip">luarocks-3.5.0-windows-64.zip</a>
                             (luarocks.exe stand-alone Windows 64-bit binary)
                        </td>
                        <td>
                            <a href="luarocks-3.5.0-windows-64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.5.0-linux-x86_64.zip">luarocks-3.5.0-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.5.0-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.5.0-win32.zip">luarocks-3.5.0-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.5.0-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.4.0.tar.gz">luarocks-3.4.0.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.4.0.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.4.0-windows-32.zip">luarocks-3.4.0-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows 32-bit binary)
                        </td>
                        <td>
                            <a href="luarocks-3.4.0-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.4.0-windows-64.zip">luarocks-3.4.0-windows-64.zip</a>
                             (luarocks.exe stand-alone Windows 64-bit binary)
                        </td>
                        <td>
                            <a href="luarocks-3.4.0-windows-64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.4.0-linux-x86_64.zip">luarocks-3.4.0-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.4.0-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.4.0-win32.zip">luarocks-3.4.0-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.4.0-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.3.1.tar.gz">luarocks-3.3.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.3.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.3.1-windows-32.zip">luarocks-3.3.1-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows 32-bit binary)
                        </td>
                        <td>
                            <a href="luarocks-3.3.1-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.3.1-windows-64.zip">luarocks-3.3.1-windows-64.zip</a>
                             (luarocks.exe stand-alone Windows 64-bit binary)
                        </td>
                        <td>
                            <a href="luarocks-3.3.1-windows-64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.3.1-linux-x86_64.zip">luarocks-3.3.1-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.3.1-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.3.1-win32.zip">luarocks-3.3.1-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.3.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.3.0.tar.gz">luarocks-3.3.0.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.3.0.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.3.0-windows-32.zip">luarocks-3.3.0-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows 32-bit binary)
                        </td>
                        <td>
                            <a href="luarocks-3.3.0-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.3.0-windows-64.zip">luarocks-3.3.0-windows-64.zip</a>
                             (luarocks.exe stand-alone Windows 64-bit binary)
                        </td>
                        <td>
                            <a href="luarocks-3.3.0-windows-64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.3.0-linux-x86_64.zip">luarocks-3.3.0-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.3.0-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.3.0-win32.zip">luarocks-3.3.0-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.3.0-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.2.1.tar.gz">luarocks-3.2.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.2.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.2.1-windows-32.zip">luarocks-3.2.1-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.2.1-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.2.1-linux-x86_64.zip">luarocks-3.2.1-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.2.1-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.2.1-win32.zip">luarocks-3.2.1-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.2.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.2.0.tar.gz">luarocks-3.2.0.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.2.0.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.2.0-windows-32.zip">luarocks-3.2.0-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.2.0-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.2.0-linux-x86_64.zip">luarocks-3.2.0-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.2.0-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.2.0-win32.zip">luarocks-3.2.0-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.2.0-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.1.3.tar.gz">luarocks-3.1.3.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.1.3.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.1.3-windows-32.zip">luarocks-3.1.3-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.1.3-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.1.3-linux-x86_64.zip">luarocks-3.1.3-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.1.3-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.1.3-win32.zip">luarocks-3.1.3-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.1.3-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.1.2.tar.gz">luarocks-3.1.2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.1.2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.1.2-windows-32.zip">luarocks-3.1.2-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.1.2-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.1.2-linux-x86_64.zip">luarocks-3.1.2-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.1.2-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.1.2-win32.zip">luarocks-3.1.2-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.1.2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.1.1.tar.gz">luarocks-3.1.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.1.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.1.1-windows-32.zip">luarocks-3.1.1-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.1.1-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.1.1-linux-x86_64.zip">luarocks-3.1.1-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.1.1-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.1.1-win32.zip">luarocks-3.1.1-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.1.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.1.0.tar.gz">luarocks-3.1.0.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.1.0.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.1.0-windows-32.zip">luarocks-3.1.0-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.1.0-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.1.0-linux-x86_64.zip">luarocks-3.1.0-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.1.0-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.1.0-win32.zip">luarocks-3.1.0-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.1.0-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.0.4.tar.gz">luarocks-3.0.4.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.0.4.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.4-windows-32.zip">luarocks-3.0.4-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.0.4-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.4-linux-x86_64.zip">luarocks-3.0.4-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.0.4-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.4-win32.zip">luarocks-3.0.4-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.0.4-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.0.3.tar.gz">luarocks-3.0.3.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.0.3.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.3-windows-32.zip">luarocks-3.0.3-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.0.3-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.3-linux-x86_64.zip">luarocks-3.0.3-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.0.3-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.3-win32.zip">luarocks-3.0.3-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.0.3-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.0.2.tar.gz">luarocks-3.0.2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.0.2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.2-windows-32.zip">luarocks-3.0.2-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.0.2-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.2-linux-x86_64.zip">luarocks-3.0.2-linux-x86_64.zip</a>
                             (luarocks stand-alone Linux x86_64 binary)
                        </td>
                        <td>
                            <a href="luarocks-3.0.2-linux-x86_64.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.2-win32.zip">luarocks-3.0.2-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.0.2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.0.1.tar.gz">luarocks-3.0.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.0.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.1-windows-32.zip">luarocks-3.0.1-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.0.1-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.1-win32.zip">luarocks-3.0.1-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.0.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.0.1-rc2.tar.gz">luarocks-3.0.1-rc2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.0.1-rc2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.1-rc2-windows-32.zip">luarocks-3.0.1-rc2-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.0.1-rc2-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.1-rc2-win32.zip">luarocks-3.0.1-rc2-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.0.1-rc2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.0.1-rc1.tar.gz">luarocks-3.0.1-rc1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.0.1-rc1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.1-rc1-windows-32.zip">luarocks-3.0.1-rc1-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.0.1-rc1-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.1-rc1-win32.zip">luarocks-3.0.1-rc1-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.0.1-rc1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.0.0.tar.gz">luarocks-3.0.0.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.0.0.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.0-windows-32.zip">luarocks-3.0.0-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.0.0-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.0-win32.zip">luarocks-3.0.0-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.0.0-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.0.0-rc2.tar.gz">luarocks-3.0.0-rc2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.0.0-rc2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.0-rc2-windows-32.zip">luarocks-3.0.0-rc2-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.0.0-rc2-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.0-rc2-win32.zip">luarocks-3.0.0-rc2-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.0.0-rc2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.0.0-rc1.tar.gz">luarocks-3.0.0-rc1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.0.0-rc1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.0-rc1-windows-32.zip">luarocks-3.0.0-rc1-windows-32.zip</a>
                             (luarocks.exe stand-alone Windows binary)
                        </td>
                        <td>
                            <a href="luarocks-3.0.0-rc1-windows-32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-3.0.0-rc1-win32.zip">luarocks-3.0.0-rc1-win32.zip</a>
                             (legacy Windows package, includes Lua 5.1)
                        </td>
                        <td>
                            <a href="luarocks-3.0.0-rc1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.0.0beta2.tar.gz">luarocks-3.0.0beta2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.0.0beta2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-3.0.0beta1.tar.gz">luarocks-3.0.0beta1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-3.0.0beta1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.4.4.tar.gz">luarocks-2.4.4.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.4.4.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.4.4-win32.zip">luarocks-2.4.4-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.4.4-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.4.3.tar.gz">luarocks-2.4.3.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.4.3.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.4.3-win32.zip">luarocks-2.4.3-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.4.3-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.4.2.tar.gz">luarocks-2.4.2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.4.2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.4.2-win32.zip">luarocks-2.4.2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.4.2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.4.1.tar.gz">luarocks-2.4.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.4.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.4.1-win32.zip">luarocks-2.4.1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.4.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.4.0.tar.gz">luarocks-2.4.0.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.4.0.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.4.0-win32.zip">luarocks-2.4.0-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.4.0-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.3.0.tar.gz">luarocks-2.3.0.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.3.0.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.3.0-win32.zip">luarocks-2.3.0-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.3.0-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.3.0-rc2.tar.gz">luarocks-2.3.0-rc2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.3.0-rc2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.3.0-rc2-win32.zip">luarocks-2.3.0-rc2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.3.0-rc2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.3.0-rc1.tar.gz">luarocks-2.3.0-rc1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.3.0-rc1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.3.0-rc1-win32.zip">luarocks-2.3.0-rc1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.3.0-rc1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.2.3-rc2.tar.gz">luarocks-2.2.3-rc2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.2.3-rc2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.2.3-rc2-win32.zip">luarocks-2.2.3-rc2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.2.3-rc2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.2.3-rc1.tar.gz">luarocks-2.2.3-rc1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.2.3-rc1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.2.3-rc1-win32.zip">luarocks-2.2.3-rc1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.2.3-rc1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.2.2.tar.gz">luarocks-2.2.2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.2.2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.2.2-win32.zip">luarocks-2.2.2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.2.2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.2.1.tar.gz">luarocks-2.2.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.2.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.2.1-win32.zip">luarocks-2.2.1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.2.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.2.0.tar.gz">luarocks-2.2.0.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.2.0.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.2.0-win32.zip">luarocks-2.2.0-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.2.0-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.2.0beta1.tar.gz">luarocks-2.2.0beta1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.2.0beta1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.2.0beta1-win32.zip">luarocks-2.2.0beta1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.2.0beta1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.1.2.tar.gz">luarocks-2.1.2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.1.2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.1.2-win32.zip">luarocks-2.1.2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.1.2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.1.1.tar.gz">luarocks-2.1.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.1.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.1.1-win32.zip">luarocks-2.1.1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.1.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.1.0.tar.gz">luarocks-2.1.0.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.1.0.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.1.0-win32.zip">luarocks-2.1.0-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.1.0-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.1.0-rc3.tar.gz">luarocks-2.1.0-rc3.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.1.0-rc3.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.1.0-rc3-win32.zip">luarocks-2.1.0-rc3-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.1.0-rc3-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.1.0-rc2.tar.gz">luarocks-2.1.0-rc2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.1.0-rc2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.1.0-rc2-win32.zip">luarocks-2.1.0-rc2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.1.0-rc2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.1.0-rc1.tar.gz">luarocks-2.1.0-rc1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.1.0-rc1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.1.0-rc1-win32.zip">luarocks-2.1.0-rc1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.1.0-rc1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.13.tar.gz">luarocks-2.0.13.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.13.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.13-win32.zip">luarocks-2.0.13-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.13-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.12.tar.gz">luarocks-2.0.12.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.12.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.12-win32.zip">luarocks-2.0.12-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.12-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.11.tar.gz">luarocks-2.0.11.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.11.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.11-win32.zip">luarocks-2.0.11-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.11-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.10.tar.gz">luarocks-2.0.10.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.10.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.10-win32.zip">luarocks-2.0.10-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.10-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.9.1.tar.gz">luarocks-2.0.9.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.9.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.9.1-win32.zip">luarocks-2.0.9.1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.9.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.9.tar.gz">luarocks-2.0.9.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.9.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.9-win32.zip">luarocks-2.0.9-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.9-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.9-rc2.tar.gz">luarocks-2.0.9-rc2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.9-rc2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.9-rc2-win32.zip">luarocks-2.0.9-rc2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.9-rc2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.9-rc1.tar.gz">luarocks-2.0.9-rc1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.9-rc1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.9-rc1-win32.zip">luarocks-2.0.9-rc1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.9-rc1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.8.tar.gz">luarocks-2.0.8.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.8.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.8-win32.zip">luarocks-2.0.8-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.8-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.8-rc2.tar.gz">luarocks-2.0.8-rc2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.8-rc2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.8-rc2-win32.zip">luarocks-2.0.8-rc2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.8-rc2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.8-rc1.tar.gz">luarocks-2.0.8-rc1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.8-rc1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.8-rc1-win32.zip">luarocks-2.0.8-rc1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.8-rc1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.7.1.tar.gz">luarocks-2.0.7.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.7.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.7.1-win32.zip">luarocks-2.0.7.1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.7.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.7.tar.gz">luarocks-2.0.7.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.7.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.7-win32.zip">luarocks-2.0.7-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.7-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.6.tar.gz">luarocks-2.0.6.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.6.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.6-win32.zip">luarocks-2.0.6-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.6-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.6-rc1.tar.gz">luarocks-2.0.6-rc1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.6-rc1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.6-rc1-win32.zip">luarocks-2.0.6-rc1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.6-rc1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.5.tar.gz">luarocks-2.0.5.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.5.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.5-win32.zip">luarocks-2.0.5-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.5-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.5-rc1.tar.gz">luarocks-2.0.5-rc1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.5-rc1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.5-rc1-win32.zip">luarocks-2.0.5-rc1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.5-rc1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.4.1.tar.gz">luarocks-2.0.4.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.4.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.4.1-win32.zip">luarocks-2.0.4.1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.4.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.4.tar.gz">luarocks-2.0.4.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.4.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.4-win32.zip">luarocks-2.0.4-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.4-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.4-rc3.tar.gz">luarocks-2.0.4-rc3.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.4-rc3.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.4-rc3-win32.zip">luarocks-2.0.4-rc3-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.4-rc3-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.4-rc2.tar.gz">luarocks-2.0.4-rc2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.4-rc2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.4-rc2-win32.zip">luarocks-2.0.4-rc2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.4-rc2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.4-rc1.tar.gz">luarocks-2.0.4-rc1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.4-rc1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.4-rc1-win32.zip">luarocks-2.0.4-rc1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.4-rc1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.3.tar.gz">luarocks-2.0.3.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.3.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.3-win32.zip">luarocks-2.0.3-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.3-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.3-rc1-win32.zip">luarocks-2.0.3-rc1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.3-rc1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.3-rc2.tar.gz">luarocks-2.0.3-rc2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.3-rc2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.3-r2-win32.zip">luarocks-2.0.3-r2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.3-r2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.3-rc1.tar.gz">luarocks-2.0.3-rc1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.3-rc1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.2.tar.gz">luarocks-2.0.2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.2-win32.zip">luarocks-2.0.2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-2.0.2-lfw.zip">luarocks-2.0.2-lfw.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.2-lfw.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.1.tar.gz">luarocks-2.0.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-2.0.tar.gz">luarocks-2.0.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-2.0.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-1.0.1.tar.gz">luarocks-1.0.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-1.0.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-1.0.1-win32.zip">luarocks-1.0.1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-1.0.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-1.0.tar.gz">luarocks-1.0.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-1.0.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-1.0-win32.zip">luarocks-1.0-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-1.0-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.6.0.2.tar.gz">luarocks-0.6.0.2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.6.0.2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-0.6.0.2-win32.zip">luarocks-0.6.0.2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-0.6.0.2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.5.2.tar.gz">luarocks-0.5.2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.5.2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-0.5.2-win32.zip">luarocks-0.5.2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-0.5.2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.5.1.tar.gz">luarocks-0.5.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.5.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-0.5.1-win32.zip">luarocks-0.5.1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-0.5.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.5.tar.gz">luarocks-0.5.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.5.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-0.5-win32.zip">luarocks-0.5-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-0.5-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.4.3.tar.gz">luarocks-0.4.3.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.4.3.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-0.4.3-win32.zip">luarocks-0.4.3-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-0.4.3-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.4.2.tar.gz">luarocks-0.4.2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.4.2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-0.4.2-win32.zip">luarocks-0.4.2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-0.4.2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.4.1.tar.gz">luarocks-0.4.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.4.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-0.4.1-win32.zip">luarocks-0.4.1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-0.4.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.4.tar.gz">luarocks-0.4.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.4.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-0.4-win32.zip">luarocks-0.4-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-0.4-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.3.2.tar.gz">luarocks-0.3.2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.3.2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-0.3.2-win32.zip">luarocks-0.3.2-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-0.3.2-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.3.1.tar.gz">luarocks-0.3.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.3.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-0.3.1-win32.zip">luarocks-0.3.1-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-0.3.1-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.3.tar.gz">luarocks-0.3.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.3.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                    <tr>
                        <td>
                            <a href="luarocks-0.3-win32.zip">luarocks-0.3-win32.zip</a>
                        </td>
                        <td>
                            <a href="luarocks-0.3-win32.zip.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.2.tar.gz">luarocks-0.2.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.2.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>

                    <tr>
                        <td>
                            <a href="luarocks-0.1.tar.gz">luarocks-0.1.tar.gz</a>
                        </td>
                        <td>
                            <a href="luarocks-0.1.tar.gz.asc">PGP signature</a>
                        </td>
                    </tr>
                </table>

            </div>
        </div>
    </div>
</body>
</html>

`
