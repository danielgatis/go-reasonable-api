import {
  Body,
  Container,
  Head,
  Html,
  Preview,
  Section,
  Tailwind,
  Text,
} from "@react-email/components";
import * as React from "react";
import { tailwindConfig } from "../tailwind.config";

// Go template placeholders
const NAME = "{{.Name}}";
const SCHEDULED_AT = "{{.ScheduledAt}}";
const DAYS_LEFT = "{{.DaysLeft}}";

export const AccountDeletionScheduled = () => {
  return (
    <Html>
      <Head />
      <Preview>Sua conta ZapAgenda sera excluida em {DAYS_LEFT} dias</Preview>
      <Tailwind config={tailwindConfig}>
        <Body className="bg-gray-100 font-sans">
          <Container className="bg-white mx-auto my-10 max-w-xl rounded-lg shadow-sm">
            <Section className="px-12 py-8 border-b border-gray-200">
              <Text className="text-2xl font-bold text-brand m-0">
                ZapAgenda
              </Text>
            </Section>

            <Section className="px-12 py-8">
              <Text className="text-xl font-bold text-red-600 mb-6">
                Exclusao de Conta Agendada
              </Text>

              <Text className="text-base text-gray-600 leading-7 mb-4">
                Ola {NAME},
              </Text>

              <Text className="text-base text-gray-600 leading-7 mb-6">
                Recebemos sua solicitacao para excluir sua conta do ZapAgenda.
                Sua conta e todos os dados associados serao permanentemente
                excluidos em <strong>{SCHEDULED_AT}</strong>.
              </Text>

              <Text className="text-base text-gray-600 leading-7 mb-6">
                Voce tem <strong>{DAYS_LEFT} dias</strong> para cancelar esta
                solicitacao. Para manter sua conta, basta fazer login no
                ZapAgenda antes da data de exclusao.
              </Text>

              <Section className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mb-6">
                <Text className="text-sm text-yellow-800 m-0">
                  <strong>Importante:</strong> Esta acao e irreversivel. Apos a
                  data de exclusao, nao sera possivel recuperar sua conta ou
                  dados.
                </Text>
              </Section>

              <Text className="text-base text-gray-600 leading-7 mb-4">
                Se voce nao solicitou a exclusao da sua conta, por favor faca
                login imediatamente para cancelar esta solicitacao e proteger
                sua conta.
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

export default AccountDeletionScheduled;
