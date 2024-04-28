package org.service.auth.chatappauthservice.exception.refresh;

public class MissingRefreshTokenException extends RuntimeException {

	public MissingRefreshTokenException(String message) {
		super(message);
	}

}
