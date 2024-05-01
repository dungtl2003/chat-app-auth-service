package org.service.auth.chatappauthservice.exception.authorize;

public class MissingAccessTokenException extends RuntimeException {

	public MissingAccessTokenException(String message) {
		super(message);
	}

}
