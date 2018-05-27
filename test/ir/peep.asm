# Test to demostrate the jumps-over-jumps peephole optimization.

	.data
n:		.word	0

	.text


	.globl main
	.ent main
main:
	li	$3, 0		# n -> $3
	# Store dirty variables back into memory
	sw	$3, n
	bne	$3, 0, elseLabel

	lw	$3, n
	addi	$3, $3, 1	# n -> $3
	# Store dirty variables back into memory
	sw	$3, n
	j	exit

elseLabel:
	# The generated assembly should not contain the label "ifLabel".
	lw	$3, n
	addi	$3, $3, 2	# n -> $3
	# Store dirty variables back into memory
	sw	$3, n

exit:
	li	$2, 1
	lw	$3, n
	move	$4, $3
	syscall
	# Store dirty variables back into memory

label1:
	# The generated assembly should not contain the label "label2".
	lw	$3, n
	addi	$3, $3, 1	# n -> $3
	# Store dirty variables back into memory
	sw	$3, n
	j	label3

label3:
	li	$2, 10
	syscall
	.end main
