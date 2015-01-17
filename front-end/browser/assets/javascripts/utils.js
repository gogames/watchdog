var _HPROSE_FUNCS	= [ 	"addServer",
    				"delServer",
              		    	"getMonitorResult",
              		    	"getUser",
              		    	"login",
              		    	"logout",
              		    	"register",
              		    	"updatePassword" ];
var _COOKIE_SID       	= "sid";
var _COOKIE_USERNAME  	= "username";
var _COOKIE_REMEMBER  	= "remember";
var _Hour 		= 1000 * 60 * 60;
var _Day 		= 24 * _Hour;
var _EXPIRE		= 14;
var client 		= new hprose.HttpClient("http://localhost:8683/", _HPROSE_FUNCS);

// cookie operations
function get_cookie_expire_date() {
	var cookie_time = new Date();
	if ($.cookie(_COOKIE_REMEMBER) != null) {
		cookie_time.setTime(cookie_time.getTime() + _Day);
	} else {
		cookie_time.setTime(cookie_time.getTime() + _Hour);
	}
	return cookie_time;
}

function set_cookie_remember() {
	$.cookie(_COOKIE_REMEMBER, null);
	$.cookie(_COOKIE_REMEMBER, true, {expires:_EXPIRE});
}

function set_cookie_sid(sid) {
	$.cookie(_COOKIE_SID, sid, {expires:get_cookie_expire_date()});
}

function set_cookie_username(username) {
	$.cookie(_COOKIE_USERNAME, username, {expires:_EXPIRE});
}

function del_cookie_sid() {
	$.cookie(_COOKIE_SID, null);
}

function del_cookie_username() {
	$.cookie(_COOKIE_USERNAME, null);
}

function get_cookie_sid() {
	return $.cookie(_COOKIE_SID);
}

function get_cookie_username() {
	return $.cookie(_COOKIE_USERNAME);
}

// hprose invoke 
function _H_get_user(success_handler, error_handler) {
	client.getUser(get_cookie_sid(), get_cookie_username(),
			function(result) 	{ success_handler(result); 	},
			function(name, err) 	{ error_handler(name, err); 	}
			);
}

function _H_login(username, password, success_handler, error_handler) {
	client.login(username, password,
			function(result) 	{ success_handler(result); 	},
			function(name, err) 	{ error_handler(name, err); 	}
		    );
}

function _H_register(username, password, success_handler, error_handler) {
	client.register(username, password,
			function(result) 	{ success_handler(result); 	},
			function(name, err) 	{ error_handler(name, err); 	}
		    );
}

function _H_get_monitor_result(server, success_handler, error_handler) {
	client.getMonitorResult(get_cookie_sid(), get_cookie_username(), server,
			function(result) 	{ success_handler(result); 	},
			function(name, err) 	{ error_handler(name, err); 	}
		    );
}

function _H_logout(success_handler, error_handler) {
	client.logout(get_cookie_sid(), get_cookie_username(),
			function(result) 	{ success_handler(result); 	},
			function(name, err) 	{ error_handler(name, err); 	}
		    );
}

function _H_add_server(server, success_handler, error_handler) {
	client.addServer(get_cookie_sid(), get_cookie_username(), server,
			function(result) 	{ success_handler(result); 	},
			function(name, err) 	{ error_handler(name, err); 	}
		    );
}

function _H_del_server(server, success_handler, error_handler) {
	client.delServer(get_cookie_sid(), get_cookie_username(), server,
			function(result) 	{ success_handler(result); 	},
			function(name, err) 	{ error_handler(name, err); 	}
		    );
}

function _H_update_password(old_pass, new_pass, success_handler, error_handler) {
	client.updatePassword(get_cookie_sid(), get_cookie_username(), old_pass, new_pass,
			function(result) 	{ success_handler(result); 	},
			function(name, err) 	{ error_handler(name, err); 	}
		    );
}
