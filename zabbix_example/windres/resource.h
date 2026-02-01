/*
** Copyright (C) 2001-2026 Zabbix SIA
**
** Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
** documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
** rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
** permit persons to whom the Software is furnished to do so, subject to the following conditions:
**
** The above copyright notice and this permission notice shall be included in all copies or substantial portions
** of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
** WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
** COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
** TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
**/

//{{NO_DEPENDENCIES}}
// Microsoft Developer Studio generated include file.
// Used by resource.rc
//
#ifndef _RESOURCE_H_
#define _RESOURCE_H_

#define ZBX_STR2(str)	#str
#define ZBX_STR(str)	ZBX_STR2(str)

#ifndef VER_FILEDESCRIPTION_STR
#	define VER_FILEDESCRIPTION_STR	{VER_FILEDESCRIPTION_STR}
#endif
#ifndef ZABBIX_VERSION_MAJOR
#	define ZABBIX_VERSION_MAJOR		{ZABBIX_VERSION_MAJOR}
#endif
#ifndef ZABBIX_VERSION_MINOR
#	define ZABBIX_VERSION_MINOR		{ZABBIX_VERSION_MINOR}
#endif
#ifndef ZABBIX_VERSION_PATCH
#	define ZABBIX_VERSION_PATCH 	{ZABBIX_VERSION_PATCH}
#endif
#ifndef ZABBIX_VERSION_RC
#	define ZABBIX_VERSION_RC		{ZABBIX_VERSION_RC}
#endif
#ifndef ZABBIX_VERSION_RC_NUM
#	define ZABBIX_VERSION_RC_NUM	{ZABBIX_RC_NUM}
#endif
#ifndef ZABBIX_LICENSE_YEARS
#	define ZABBIX_LICENSE_YEARS		{ZABBIX_LICENSE_YEARS}
#endif

#define VER_FILEVERSION		ZABBIX_VERSION_MAJOR,ZABBIX_VERSION_MINOR,ZABBIX_VERSION_PATCH,ZABBIX_VERSION_RC_NUM
#define VER_FILEVERSION_STR	ZBX_STR(ZABBIX_VERSION_MAJOR) "." ZBX_STR(ZABBIX_VERSION_MINOR) "." \
					ZBX_STR(ZABBIX_VERSION_PATCH) "." ZBX_STR(ZABBIX_VERSION_REVISION) "\0"
#define VER_PRODUCTVERSION	ZABBIX_VERSION_MAJOR,ZABBIX_VERSION_MINOR,ZABBIX_VERSION_PATCH
#define VER_PRODUCTVERSION_STR	ZBX_STR(ZABBIX_VERSION_MAJOR) "." ZBX_STR(ZABBIX_VERSION_MINOR) "." \
					ZBX_STR(ZABBIX_VERSION_PATCH) ZABBIX_VERSION_RC "\0"
#define VER_COMPANYNAME_STR	"Zabbix SIA\0"
#define VER_LEGALCOPYRIGHT_STR	"Copyright (C) " ZABBIX_LICENSE_YEARS " " VER_COMPANYNAME_STR
#define VER_PRODUCTNAME_STR	"Zabbix\0"

// Next default values for new objects
//
#ifdef APSTUDIO_INVOKED
#ifndef APSTUDIO_READONLY_SYMBOLS
#define _APS_NEXT_RESOURCE_VALUE	105
#define _APS_NEXT_COMMAND_VALUE		40001
#define _APS_NEXT_CONTROL_VALUE		1000
#define _APS_NEXT_SYMED_VALUE		101
#endif
#endif

#endif	/* _RESOURCE_H_ */
