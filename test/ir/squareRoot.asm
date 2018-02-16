# Test to find square root (floor value) of a number

	.data
nStr:	.asciiz "Enter n: "
x:	.word	0
ans:	.word	0
start:	.word	0
end:	.word	0
mid:	.word	0
temp:	.word	0
str:	.asciiz "Square root of n: "

	.text


	.globl main
	.ent main
main:
	li $v0, 4
	la $a0, nStr
	syscall
	li $v0, 5
	syscall
	move $t1, $v0
	move $t4, $t1		# ans -> $t4
	# Store variables back into memory
	sw $t4, ans
	sw $t1, x

	lw $t1, x
	beq $t1, 0, exit		# exit -> $t0
	# Store variables back into memory
	sw $t1, x

	lw $t1, x
	beq $t1, 1, exit		# exit -> $t0
	li $t4, 1		# start -> $t4
	move $t3, $t1		# end -> $t3
	# Store variables back into memory
	sw $t4, start
	sw $t3, end
	sw $t1, x

while:
	lw $t1, start
	lw $t4, end
	add $t3, $t1, $t4	# mid -> $t3
	srl $t3, $t3, 1		# mid -> $t3
	mul $t2, $t3, $t3	# temp -> $t2
	# x is a perfect square
	# Store variables back into memory
	sw $t1, start
	sw $t4, end
	sw $t3, mid
	sw $t2, temp

	lw $t1, temp
	lw $t4, x
	beq $t1, $t4, perfectSquare		# perfectSquare -> $t0
	# Store variables back into memory
	sw $t1, temp
	sw $t4, x

	lw $t1, temp
	lw $t4, x
	blt $t1, $t4, ifBranch		# ifBranch -> $t0
	# else branch
	lw $t3, mid
	sw $t4, x		# spilled x, freed $t4
	sub $t4, $t3, 1		# end -> $t4
	# Store variables back into memory
	sw $t1, temp
	sw $t4, end
	sw $t3, mid

	lw $t1, start
	lw $t4, end
	ble $t1, $t4, while		# while -> $t0
	# Store variables back into memory
	sw $t1, start
	sw $t4, end

	j exit

ifBranch:
	lw $t1, mid
	addi $t4, $t1, 1		# start -> $t4
	move $t3, $t1		# ans -> $t3
	# Store variables back into memory
	sw $t3, ans
	sw $t1, mid
	sw $t4, start

	lw $t1, start
	lw $t4, end
	ble $t1, $t4, while		# while -> $t0
	# Store variables back into memory
	sw $t1, start
	sw $t4, end

	j exit

perfectSquare:
	lw $t1, mid
	move $t4, $t1		# ans -> $t4
	# Store variables back into memory
	sw $t1, mid
	sw $t4, ans

exit:
	li $v0, 4
	la $a0, str
	syscall
	li $v0, 1
	lw $t1, ans
	move $a0, $t1
	syscall
	# Store variables back into memory
	sw $t1, ans
	li $v0, 10
	syscall
	.end main
