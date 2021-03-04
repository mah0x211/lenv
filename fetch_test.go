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
	exp := []string{"lua-5.4.2.tar.gz", "lua-5.4.1.tar.gz", "lua-5.4.0.tar.gz", "lua-5.3.6.tar.gz", "lua-5.3.5.tar.gz", "lua-5.3.4.tar.gz", "lua-5.3.3.tar.gz", "lua-5.3.2.tar.gz", "lua-5.3.1.tar.gz", "lua-5.3.0.tar.gz", "lua-5.2.4.tar.gz", "lua-5.2.3.tar.gz", "lua-5.2.2.tar.gz", "lua-5.2.1.tar.gz", "lua-5.2.0.tar.gz", "lua-5.1.5.tar.gz", "lua-5.1.4.tar.gz", "lua-5.1.3.tar.gz", "lua-5.1.2.tar.gz", "lua-5.1.1.tar.gz", "lua-5.1.tar.gz", "lua-5.0.3.tar.gz", "lua-5.0.2.tar.gz", "lua-5.0.1.tar.gz", "lua-5.0.tar.gz", "lua-4.0.1.tar.gz", "lua-4.0.tar.gz", "lua-3.2.2.tar.gz", "lua-3.2.1.tar.gz", "lua-3.2.tar.gz", "lua-3.1.tar.gz", "lua-3.0.tar.gz", "lua-2.5.tar.gz", "lua-2.4.tar.gz", "lua-2.2.tar.gz", "lua-2.1.tar.gz", "lua-1.1.tar.gz", "lua-1.0.tar.gz"}
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
<TH CLASS="size">size</TH>
<TH>checksums</TH>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.4.2.tar.gz">lua-5.4.2.tar.gz</A></TD>
<TD CLASS="date">2020-11-13</TD>
<TD CLASS="size">353472</TD>
<TD CLASS="sum">md5: 49c92d6a49faba342c35c52e1ac3f81e<BR>sha1: 96d4a21393c94bed286b8dc0568f4bdde8730b22</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.4.1.tar.gz">lua-5.4.1.tar.gz</A></TD>
<TD CLASS="date">2020-09-30</TD>
<TD CLASS="size">353965</TD>
<TD CLASS="sum">md5: 1d575faef1c907292edd79e7a2784d30<BR>sha1: 88961e7d4fda58ca2c6163938fd48db8880e803d</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.4.0.tar.gz">lua-5.4.0.tar.gz</A></TD>
<TD CLASS="date">2020-06-18</TD>
<TD CLASS="size">349308</TD>
<TD CLASS="sum">md5: dbf155764e5d433fc55ae80ea7060b60<BR>sha1: 8cdbffa8a214a23d190d7c45f38c19518ae62e89</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.6.tar.gz">lua-5.3.6.tar.gz</A></TD>
<TD CLASS="date">2020-09-14</TD>
<TD CLASS="size">303770</TD>
<TD CLASS="sum">md5: 83f23dbd5230140a3770d5f54076948d<BR>sha1: f27d20d6c81292149bc4308525a9d6733c224fa5</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.5.tar.gz">lua-5.3.5.tar.gz</A></TD>
<TD CLASS="date">2018-06-26</TD>
<TD CLASS="size">303543</TD>
<TD CLASS="sum">md5: 4f4b4f323fd3514a68e0ab3da8ce3455<BR>sha1: 112eb10ff04d1b4c9898e121d6bdf54a81482447</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.4.tar.gz">lua-5.3.4.tar.gz</A></TD>
<TD CLASS="date">2017-01-12</TD>
<TD CLASS="size">303586</TD>
<TD CLASS="sum">md5: 53a9c68bcc0eda58bdc2095ad5cdfc63<BR>sha1: 79790cfd40e09ba796b01a571d4d63b52b1cd950</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.3.tar.gz">lua-5.3.3.tar.gz</A></TD>
<TD CLASS="date">2016-05-30</TD>
<TD CLASS="size">294290</TD>
<TD CLASS="sum">md5: 703f75caa4fdf4a911c1a72e67a27498<BR>sha1: a0341bc3d1415b814cc738b2ec01ae56045d64ef</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.2.tar.gz">lua-5.3.2.tar.gz</A></TD>
<TD CLASS="date">2015-11-25</TD>
<TD CLASS="size">288235</TD>
<TD CLASS="sum">md5: 33278c2ab5ee3c1a875be8d55c1ca2a1<BR>sha1: 7a47adef554fdca7d0c5536148de34579134a973</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.1.tar.gz">lua-5.3.1.tar.gz</A></TD>
<TD CLASS="date">2015-06-10</TD>
<TD CLASS="size">282401</TD>
<TD CLASS="sum">md5: 797adacada8d85761c079390ff1d9961<BR>sha1: 1676c6a041d90b6982db8cef1e5fb26000ab6dee</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.3.0.tar.gz">lua-5.3.0.tar.gz</A></TD>
<TD CLASS="date">2015-01-06</TD>
<TD CLASS="size">278045</TD>
<TD CLASS="sum">md5: a1b0a7e92d0c85bbff7a8d27bf29f8af<BR>sha1: 1c46d1c78c44039939e820126b86a6ae12dadfba</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.2.4.tar.gz">lua-5.2.4.tar.gz</A></TD>
<TD CLASS="date">2015-02-26</TD>
<TD CLASS="size">252651</TD>
<TD CLASS="sum">md5: 913fdb32207046b273fdb17aad70be13<BR>sha1: ef15259421197e3d85b7d6e4871b8c26fd82c1cf</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.2.3.tar.gz">lua-5.2.3.tar.gz</A></TD>
<TD CLASS="date">2013-11-11</TD>
<TD CLASS="size">251195</TD>
<TD CLASS="sum">md5: dc7f94ec6ff15c985d2d6ad0f1b35654<BR>sha1: 926b7907bc8d274e063d42804666b40a3f3c124c</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.2.2.tar.gz">lua-5.2.2.tar.gz</A></TD>
<TD CLASS="date">2013-03-21</TD>
<TD CLASS="size">251713</TD>
<TD CLASS="sum">md5: efbb645e897eae37cad4344ce8b0a614<BR>sha1: 0857e41e5579726a4cb96732e80d7aa47165eaf5</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.2.1.tar.gz">lua-5.2.1.tar.gz</A></TD>
<TD CLASS="date">2012-06-08</TD>
<TD CLASS="size">249882</TD>
<TD CLASS="sum">md5: ae08f641b45d737d12d30291a5e5f6e3<BR>sha1: 6bb1b0a39b6a5484b71a83323c690154f86b2021</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.2.0.tar.gz">lua-5.2.0.tar.gz</A></TD>
<TD CLASS="date">2011-12-12</TD>
<TD CLASS="size">246377</TD>
<TD CLASS="sum">md5: f1ea831f397214bae8a265995ab1a93e<BR>sha1: 08f84c355cdd646f617f09cebea48bd832415829</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.1.5.tar.gz">lua-5.1.5.tar.gz</A></TD>
<TD CLASS="date">2012-02-13</TD>
<TD CLASS="size">221213</TD>
<TD CLASS="sum">md5: 2e115fe26e435e33b0d5c022e4490567<BR>sha1: b3882111ad02ecc6b972f8c1241647905cb2e3fc</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.1.4.tar.gz">lua-5.1.4.tar.gz</A></TD>
<TD CLASS="date">2008-08-18</TD>
<TD CLASS="size">216679</TD>
<TD CLASS="sum">md5: d0870f2de55d59c1c8419f36e8fac150<BR>sha1: 2b11c8e60306efb7f0734b747588f57995493db7</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.1.3.tar.gz">lua-5.1.3.tar.gz</A></TD>
<TD CLASS="date">2008-01-21</TD>
<TD CLASS="size">215817</TD>
<TD CLASS="sum">md5: a70a8dfaa150e047866dc01a46272599<BR>sha1: 89bc9f5a351402565b8077e8123327e7cd15f004</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.1.2.tar.gz">lua-5.1.2.tar.gz</A></TD>
<TD CLASS="date">2007-03-29</TD>
<TD CLASS="size">214134</TD>
<TD CLASS="sum">md5: 687ce4c2a1ddff18f1008490fdc4e5e0<BR>sha1: 8a460d2d7e70e93cb72bf3d584405464763cb5f0</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.1.1.tar.gz">lua-5.1.1.tar.gz</A></TD>
<TD CLASS="date">2006-06-07</TD>
<TD CLASS="size">207810</TD>
<TD CLASS="sum">md5: 22f4f912f20802c11006fe9b84d5c461<BR>sha1: be13878ceef8e1ee7a4201261f0adf09f89f1005</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.1.tar.gz">lua-5.1.tar.gz</A></TD>
<TD CLASS="date">2006-02-20</TD>
<TD CLASS="size">206877</TD>
<TD CLASS="sum">md5: 3e8dfe8be00a744cec2f9e766b2f2aee<BR>sha1: 1ae9ec317511d525c7999c842ca0b1ddde84e374</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.0.3.tar.gz">lua-5.0.3.tar.gz</A></TD>
<TD CLASS="date">2006-06-19</TD>
<TD CLASS="size">191384</TD>
<TD CLASS="sum">md5: feee27132056de2949ce499b0ef4c480<BR>sha1: e7e91f78b8a8deb09b13436829bed557a46af8ae</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.0.2.tar.gz">lua-5.0.2.tar.gz</A></TD>
<TD CLASS="date">2004-03-17</TD>
<TD CLASS="size">190442</TD>
<TD CLASS="sum">md5: dea74646b7e5c621fef7174df83c34b1<BR>sha1: a200cfd20a9a4c7da1206ae45dddf26186a9e0e7</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.0.1.tar.gz">lua-5.0.1.tar.gz</A></TD>
<TD CLASS="date">2003-11-25</TD>
<TD CLASS="size">193978</TD>
<TD CLASS="sum">md5: e0a450d84971a3f4563b98172d1e382c<BR>sha1: 03b47b4785178aca583333f01d8726a8ab9f7ae7</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-5.0.tar.gz">lua-5.0.tar.gz</A></TD>
<TD CLASS="date">2003-04-11</TD>
<TD CLASS="size">187287</TD>
<TD CLASS="sum">md5: 6f14803fad389fb1cb15d17edfeddd91<BR>sha1: 88b1bc057857c0db5ace491c4af2c917a2b803bf</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-4.0.1.tar.gz">lua-4.0.1.tar.gz</A></TD>
<TD CLASS="date">2002-07-04</TD>
<TD CLASS="size">158426</TD>
<TD CLASS="sum">md5: a31d963dbdf727f9b34eee1e0d29132c<BR>sha1: 12f1864a7ecd4b8011862a07fa3f177b2e80e7d3</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-4.0.tar.gz">lua-4.0.tar.gz</A></TD>
<TD CLASS="date">2000-11-06</TD>
<TD CLASS="size">157102</TD>
<TD CLASS="sum">md5: be11522d46d33a931868c03694aaeeef<BR>sha1: 8d432c73ef6e98b81d252114be1a83182cc9607a</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-3.2.2.tar.gz">lua-3.2.2.tar.gz</A></TD>
<TD CLASS="date">2000-02-22</TD>
<TD CLASS="size">127768</TD>
<TD CLASS="sum">md5: 374ba5c4839709922de40b8d10382705<BR>sha1: fa50ff14c00d8523c8a3d1d3f4887ecc4400d0c3</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-3.2.1.tar.gz">lua-3.2.1.tar.gz</A></TD>
<TD CLASS="date">1999-11-25</TD>
<TD CLASS="size">127644</TD>
<TD CLASS="sum">md5: 47264a1978df49fc1dea6ffcddb05b21<BR>sha1: d43af5a1c7a65c0ddb4b0ac06c29ecf4cdd22367</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-3.2.tar.gz">lua-3.2.tar.gz</A></TD>
<TD CLASS="date">1999-07-08</TD>
<TD CLASS="size">128597</TD>
<TD CLASS="sum">md5: a6552da3d40ae9b04489a788262279e8<BR>sha1: 84cf9f0e7d00eed3ea8b4ac2b84254b714510b34</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-3.1.tar.gz">lua-3.1.tar.gz</A></TD>
<TD CLASS="date">1998-07-12</TD>
<TD CLASS="size">114186</TD>
<TD CLASS="sum">md5: d677f3827167eefdefc7b211397cfdfb<BR>sha1: 509485e3baafd946f4ffe2a984f8a63746adc32a</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-3.0.tar.gz">lua-3.0.tar.gz</A></TD>
<TD CLASS="date">1997-07-01</TD>
<TD CLASS="size">99921</TD>
<TD CLASS="sum">md5: 997558ae76c2f1cd1e10fd3835c45c6a<BR>sha1: 5c8c910353f717ba29b4fe7d538994454229b335</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-2.5.tar.gz">lua-2.5.tar.gz</A></TD>
<TD CLASS="date">1996-11-21</TD>
<TD CLASS="size">185786</TD>
<TD CLASS="sum">md5: da915d58904e75b9b0fc18147e19b0bb<BR>sha1: 7920e12c40242932c22fa261ff114cc485a39d99</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-2.4.tar.gz">lua-2.4.tar.gz</A></TD>
<TD CLASS="date">1996-05-17</TD>
<TD CLASS="size">132500</TD>
<TD CLASS="sum">md5: 5d035cc244285c1dbbcaaa0908b58965<BR>sha1: 74036935b36e6ae4ed17bd7a9408154f9a4a6b17</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-2.2.tar.gz">lua-2.2.tar.gz</A></TD>
<TD CLASS="date">1995-11-28</TD>
<TD CLASS="size">108261</TD>
<TD CLASS="sum">md5: a298b58e197ff8168ec907d6145252ef<BR>sha1: 2d8b1df94b2fb76f0f16ca1ddc54d5186b10df4b</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-2.1.tar.gz">lua-2.1.tar.gz</A></TD>
<TD CLASS="date">1995-09-13</TD>
<TD CLASS="size">125334</TD>
<TD CLASS="sum">md5: 053a9f6728cc56f6a23716a6a1ede595<BR>sha1: b9a797547f480bcb58b5d3da846c8ac8d2201df0</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-1.1.tar.gz">lua-1.1.tar.gz</A></TD>
<TD CLASS="date">1995-02-02</TD>
<TD CLASS="size">158285</TD>
<TD CLASS="sum">md5: 9f83141cc8ea362497e272071eda5cf6<BR>sha1: 67209701eec5cc633e829d023fbff62d5d6c8e5e</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-1.0.tar.gz">lua-1.0.tar.gz</A></TD>
<TD CLASS="date">2003-10-10</TD>
<TD CLASS="size">33149</TD>
<TD CLASS="sum">md5: 96e8399fc508d128badd8ac3aa8f2119<BR>sha1: 6a82d2ae7ce9ad98c7b4824a325b91522c0d6ebb</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="lua-all.tar.gz">lua-all.tar.gz</A></TD>
<TD CLASS="date">2020-12-08</TD>
<TD CLASS="size">4669322</TD>
<TD CLASS="sum">md5: 9127a2bb930d86e81e8420351375603d<BR>sha1: 1fc9740430e7d1deb91fc573a1245a73d17a872f</TD>
</TR>

</TABLE>

<H2><A NAME="manuals">Manuals</A></H2>
<TABLE>
<TR>
<TH>filename</TH>
<TH>date</TH>
<TH CLASS="size">size</TH>
<TH>checksums</TH>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-5.4.tar.gz">refman-5.4.tar.gz</A></TD>
<TD CLASS="date">2020-10-10</TD>
<TD CLASS="size">440320</TD>
<TD CLASS="sum">md5: 0b33ff65c7a3516c58de5e5a5e8d02dc<BR>sha1: 6f56b4c4fae629780ae085729b46c16c67741649</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-5.3.tar.gz">refman-5.3.tar.gz</A></TD>
<TD CLASS="date">2020-09-25</TD>
<TD CLASS="size">110173</TD>
<TD CLASS="sum">md5: 90646a8f28071a0f19d592ef15f6513f<BR>sha1: c3f6e87593d3b81959bb46e32e09ac2681a0331b</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-5.2.tar.gz">refman-5.2.tar.gz</A></TD>
<TD CLASS="date">2016-07-01</TD>
<TD CLASS="size">102134</TD>
<TD CLASS="sum">md5: 41c445ccafae43b0fa5a7c6b03453bc9<BR>sha1: c1419e4d184322d66febac7ca041d8b15fc6cefe</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-5.1.tar.gz">refman-5.1.tar.gz</A></TD>
<TD CLASS="date">2016-07-01</TD>
<TD CLASS="size">86641</TD>
<TD CLASS="sum">md5: 841c4c7607a5b853ea53b30171faf328<BR>sha1: 1d6656a0171f5697582870f7e02b5f9df4aa8e6c</TD>
</TR>


<TR>
<TD CLASS="name"><A HREF="refman-5.0.pdf">refman-5.0.pdf</A></TD>
<TD CLASS="date">2003-11-25</TD>
<TD CLASS="size">474818</TD>
<TD CLASS="sum">md5: b6eb330e98446f2040cb61ee557b24ea<BR>sha1: 21b4ee382b588f111b7789e46da1d5d797b945f3</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-5.0.ps.gz">refman-5.0.ps.gz</A></TD>
<TD CLASS="date">2003-11-25</TD>
<TD CLASS="size">224198</TD>
<TD CLASS="sum">md5: 4b0cedef4880bf925da9537520d93b57<BR>sha1: 9fa9faac3a5dc1c4479175a085d3f4ef48189260</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-4.0.pdf">refman-4.0.pdf</A></TD>
<TD CLASS="date">2000-11-07</TD>
<TD CLASS="size">446955</TD>
<TD CLASS="sum">md5: fb36fbb993e70c6f71d359d95ed87129<BR>sha1: e6e7bef7794b49289d938af87385356b2511d926</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-4.0.ps.gz">refman-4.0.ps.gz</A></TD>
<TD CLASS="date">2000-11-06</TD>
<TD CLASS="size">156876</TD>
<TD CLASS="sum">md5: 5454698095c45917ce80c934066cb76c<BR>sha1: 9191b371814f93fe5223c443ca44cb73b9c45cd4</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-3.2.pdf">refman-3.2.pdf</A></TD>
<TD CLASS="date">1999-07-08</TD>
<TD CLASS="size">321348</TD>
<TD CLASS="sum">md5: fe012393b3bd11e069231a21f5ffdd20<BR>sha1: 9c9a86d2145c52fd705d3487c06c7d4eaf7bf2dd</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-3.2.ps.gz">refman-3.2.ps.gz</A></TD>
<TD CLASS="date">1999-07-02</TD>
<TD CLASS="size">136552</TD>
<TD CLASS="sum">md5: 5b5a0b716194fcbc374ed9aac3c078e3<BR>sha1: e438b5b53d2f05f28e5f1e1f0c62d4a6da0aaac0</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-3.1.ps.gz">refman-3.1.ps.gz</A></TD>
<TD CLASS="date">1998-07-01</TD>
<TD CLASS="size">129519</TD>
<TD CLASS="sum">md5: 852e53df287708c8335f3de37e5856a3<BR>sha1: 5a441993778e4dd6c0c5e34536c78fd40c3449b8</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-2.5.ps.gz">refman-2.5.ps.gz</A></TD>
<TD CLASS="date">1996-11-18</TD>
<TD CLASS="size">115860</TD>
<TD CLASS="sum">md5: bec74e774059051f88525078f2b1c347<BR>sha1: 3237be1d8e0e8a15f2b4d7e8d55b4d61378c1d09</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-2.4.ps.gz">refman-2.4.ps.gz</A></TD>
<TD CLASS="date">1996-05-14</TD>
<TD CLASS="size">69189</TD>
<TD CLASS="sum">md5: f4e7508b4c4685e7fa8e33a49f67b344<BR>sha1: ca609ac0f5fc7f70cd98f1309aa0c30cb210ad72</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-2.2.ps.gz">refman-2.2.ps.gz</A></TD>
<TD CLASS="date">1995-11-28</TD>
<TD CLASS="size">56724</TD>
<TD CLASS="sum">md5: b1f2a2c7cfdae338bfba9314f5c8cc12<BR>sha1: c20a72f4f1336e5d0811b0b8f75fad26a64c6e5c</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-2.1.ps.gz">refman-2.1.ps.gz</A></TD>
<TD CLASS="date">1995-02-08</TD>
<TD CLASS="size">55171</TD>
<TD CLASS="sum">md5: 5d526b9b9fa4566feb18b3f0836abbc7<BR>sha1: 763402fa89946d3cf0f2707566c008cad7e58fbc</TD>
</TR>

<TR>
<TD CLASS="name"><A HREF="refman-1.1.ps.gz">refman-1.1.ps.gz</A></TD>
<TD CLASS="date">1994-05-27</TD>
<TD CLASS="size">45131</TD>
<TD CLASS="sum">md5: eecb396a07098c93c82a271346d33cc3<BR>sha1: 58d1d980296138b8a33723b1749b633b551bb5c7</TD>
</TR>

</TABLE>

<P CLASS="footer">
Last update:
Tue Dec  8 10:29:13 UTC 2020
</P>
<!--
Last change: Lua 5.4.2 released
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
