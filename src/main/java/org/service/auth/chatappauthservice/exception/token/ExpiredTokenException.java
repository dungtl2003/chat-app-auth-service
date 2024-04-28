package org.service.auth.chatappauthservice.exception.token;

public class ExpiredTokenException extends RuntimeException {

	public ExpiredTokenException(String message) {
		super(message);
	}

}
