package org.service.auth.chatappauthservice.response;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class AuthorizationResponse {

	@JsonProperty("message")
	String message;

}
