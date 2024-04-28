package org.service.auth.chatappauthservice.utils;

import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseCookie;

import java.util.ArrayList;
import java.util.List;

public class Helper {

	public static void clearCookies(HttpHeaders headers, String... cookieNames) {
		List<String> cookies = new ArrayList<>();
		for (String name : cookieNames) {
			cookies.add(ResponseCookie.from(name, "").maxAge(0).build().toString());
		}

		headers.put(HttpHeaders.SET_COOKIE, cookies);
	}

}
