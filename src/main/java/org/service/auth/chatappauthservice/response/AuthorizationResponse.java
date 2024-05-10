package org.service.auth.chatappauthservice.response;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.Builder;
import lombok.Data;

@Data
@Builder
public class AuthorizationResponse {

	@JsonProperty("message")
	String message;

	@JsonProperty("pod_ip")
	@JsonInclude(value = JsonInclude.Include.NON_EMPTY, content = JsonInclude.Include.NON_NULL)
	String podIP;

}
