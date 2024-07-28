package org.service.auth.chatappauthservice.exception.refresh;

public class ReusedRefreshTokenException extends RuntimeException {

	public ReusedRefreshTokenException(String message) {
		super(message);
	}

}
