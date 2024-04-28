package org.service.auth.chatappauthservice.exception.authorize;

public class InvalidAuthorizationHeaderException extends RuntimeException {

	public InvalidAuthorizationHeaderException(String message) {
		super(message);
	}

}
