package org.service.auth.chatappauthservice.response;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class LogoutResponse {

	// because spring boot doesn't allow return object with 0 getter method
	@JsonProperty("message")
	private String message;

}