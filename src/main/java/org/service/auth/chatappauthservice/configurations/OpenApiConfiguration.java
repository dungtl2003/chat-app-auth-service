package org.service.auth.chatappauthservice.configurations;

import io.netty.handler.codec.http.HttpScheme;
import io.swagger.v3.oas.annotations.OpenAPIDefinition;
import io.swagger.v3.oas.annotations.enums.ParameterIn;
import io.swagger.v3.oas.annotations.enums.SecuritySchemeIn;
import io.swagger.v3.oas.annotations.enums.SecuritySchemeType;
import io.swagger.v3.oas.annotations.info.Contact;
import io.swagger.v3.oas.annotations.info.Info;
import io.swagger.v3.oas.annotations.info.License;
import io.swagger.v3.oas.annotations.security.SecurityScheme;
import io.swagger.v3.oas.annotations.security.SecuritySchemes;
import io.swagger.v3.oas.annotations.servers.Server;
import org.springframework.context.annotation.Configuration;

@OpenAPIDefinition(
		info = @Info(contact = @Contact(name = "Dung", email = "dungtl2003@gmail.com"),
				description = "OpenAPI documentation for Chat-app-auth-service", title = "OpenAPI document",
				version = "1.0", license = @License(name = "", url = ""), termsOfService = "Terms of service"),
		servers = { @Server(description = "Local ENV", url = "http://localhost:8080"),
				@Server(description = "PROD ENV", url = "/") })
@SecuritySchemes(value = { @SecurityScheme(name = "authorization",
		description = "Enter the token with the `Bearer ` prefix, e.g. \"Bearer abcde12345\"", scheme = "bearer",
		in = SecuritySchemeIn.HEADER, type = SecuritySchemeType.APIKEY, bearerFormat = "JWT"),
		@SecurityScheme(name = "refresh_token", description = "Enter the refresh_token", in = SecuritySchemeIn.COOKIE,
				type = SecuritySchemeType.APIKEY) })
@Configuration
public class OpenApiConfiguration {

}
