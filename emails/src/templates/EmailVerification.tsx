import {
  Body,
  Button,
  Container,
  Head,
  Html,
  Link,
  Preview,
  Section,
  Tailwind,
  Text,
} from "@react-email/components";
import * as React from "react";
import { tailwindConfig } from "../tailwind.config";

// Go template placeholders
const NAME = "{{.Name}}";
const VERIFICATION_LINK = "{{.VerificationLink}}";

export const EmailVerification = () => {
  return (
    <Html>
      <Head />
      <Preview>Confirme seu email no ZapAgenda</Preview>
      <Tailwind config={tailwindConfig}>
        <Body className="bg-gray-100 font-sans">
          <Container className="bg-white mx-auto my-10 max-w-xl rounded-lg shadow-sm">
            <Section className="px-12 py-8 border-b border-gray-200">
              <Text className="text-2xl font-bold text-brand m-0">
                ZapAgenda
              </Text>
            </Section>

            <Section className="px-12 py-8">
              <Text className="text-xl font-bold text-brand mb-6">
                Confirme seu Email
              </Text>

              <Text className="text-base text-gray-600 leading-7 mb-4">
                Olá {NAME},
              </Text>

              <Text className="text-base text-gray-600 leading-7 mb-6">
                Obrigado por se cadastrar no ZapAgenda! Para completar seu
                cadastro e garantir a segurança da sua conta, por favor confirme
                seu email clicando no botão abaixo:
              </Text>

              <Button
                href={VERIFICATION_LINK}
                className="bg-brand text-white font-semibold py-3 px-6 rounded-lg"
              >
                Confirmar Email
              </Button>

              <Text className="text-base text-gray-600 leading-7 mt-6 mb-4">
                Se você não criou uma conta no ZapAgenda, pode ignorar este
                email com segurança.
              </Text>

              <Text className="text-base text-gray-600 leading-7 mb-4">
                Este link expira em 24 horas. Se precisar de um novo link,
                acesse sua conta e solicite um novo email de confirmação.
              </Text>

              <Text className="text-sm text-gray-400 mt-8">
                Se o botão não funcionar, copie e cole este link no seu
                navegador:
                <br />
                <Link href={VERIFICATION_LINK} className="text-brand break-all">
                  {VERIFICATION_LINK}
                </Link>
              </Text>
            </Section>

            <Section className="px-12 py-6 border-t border-gray-200">
              <Text className="text-xs text-gray-400 text-center m-0">
                © {new Date().getFullYear()} ZapAgenda. Todos os direitos
                reservados.
              </Text>
            </Section>
          </Container>
        </Body>
      </Tailwind>
    </Html>
  );
};

export default EmailVerification;
